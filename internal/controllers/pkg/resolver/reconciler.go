/*
Copyright 2021 NDD.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resolver

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"
	pkgmetav1 "github.com/yndd/ndd-core/apis/pkg/meta/v1"
	v1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/yndd/ndd-core/internal/dag"
	"github.com/yndd/ndd-core/internal/nddpkg"
	"github.com/yndd/ndd-runtime/pkg/event"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	reconcileTimeout = 1 * time.Minute

	shortWait = 30 * time.Second

	packageTagFmt = "%s:%s"
)

const (
	finalizer = "lock.pkg.ndd.yndd.io"

	errGetLock              = "cannot get package lock"
	errAddFinalizer         = "cannot add lock finalizer"
	errBuildDAG             = "cannot build DAG"
	errSortDAG              = "cannot sort DAG"
	errMissingDependencyFmt = "missing package (%s) is not a dependency"
	errInvalidConstraint    = "version constraint on dependency is invalid"
	errInvalidDependency    = "dependency package is not valid"
	errFetchTags            = "cannot fetch dependency package tags"
	errNoValidVersion       = "cannot find a valid version for package constraints"
	errNoValidVersionFmt    = "dependency (%s) does not have version in constraints (%s)"
	errInvalidPackageType   = "cannot create invalid package dependency type"
	errCreateDependency     = "cannot create dependency package"
)

// ReconcilerOption is used to configure the Reconciler.
type ReconcilerOption func(*Reconciler)

// WithLogger specifies how the Reconciler should log messages.
func WithLogger(log logging.Logger) ReconcilerOption {
	return func(r *Reconciler) {
		r.log = log
	}
}

// WithRecorder specifies how the Reconciler should record Kubernetes events.
func WithRecorder(er event.Recorder) ReconcilerOption {
	return func(r *Reconciler) {
		r.record = er
	}
}

// WithFinalizer specifies how the Reconciler should finalize package revisions.
func WithFinalizer(f resource.Finalizer) ReconcilerOption {
	return func(r *Reconciler) {
		r.lock = f
	}
}

// WithNewDagFn specifies how the Reconciler should build its dependency graph.
func WithNewDagFn(f dag.NewDAGFn) ReconcilerOption {
	return func(r *Reconciler) {
		r.newDag = f
	}
}

// WithFetcher specifies how the Reconciler should fetch package tags.
func WithFetcher(f nddpkg.Fetcher) ReconcilerOption {
	return func(r *Reconciler) {
		r.fetcher = f
	}
}

// Reconciler reconciles packages.
type Reconciler struct {
	client  client.Client
	log     logging.Logger
	record  event.Recorder
	lock    resource.Finalizer
	newDag  dag.NewDAGFn
	fetcher nddpkg.Fetcher
}

// Setup adds a controller that reconciles the Lock.
func Setup(mgr ctrl.Manager, l logging.Logger, namespace string) error {
	name := "packages/" + strings.ToLower(v1.LockGroupKind)

	clientset, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		return errors.Wrap(err, "failed to initialize clientset")
	}

	r := NewReconciler(mgr,
		WithLogger(l.WithValues("controller", name)),
		WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		WithFetcher(nddpkg.NewK8sFetcher(clientset, namespace)),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1.Lock{}).
		Owns(&v1.ProviderRevision{}).
		Complete(r)
}

// NewReconciler creates a new package revision reconciler.
func NewReconciler(mgr manager.Manager, opts ...ReconcilerOption) *Reconciler {
	r := &Reconciler{
		client:  mgr.GetClient(),
		lock:    resource.NewAPIFinalizer(mgr.GetClient(), finalizer),
		log:     logging.NewNopLogger(),
		record:  event.NewNopRecorder(),
		newDag:  dag.NewMapDag,
		fetcher: nddpkg.NewNopFetcher(),
	}

	for _, f := range opts {
		f(r)
	}

	return r
}

// Reconcile package revision.
func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) { // nolint:gocyclo
	log := r.log.WithValues("request", req)
	log.Debug("Resolver Reconciling Package Revision", "NameSpace", req.NamespacedName)

	ctx, cancel := context.WithTimeout(ctx, reconcileTimeout)
	defer cancel()

	lock := &v1.Lock{}
	if err := r.client.Get(ctx, req.NamespacedName, lock); err != nil {
		// There's no need to requeue if we no longer exist. Otherwise we'll be
		// requeued implicitly because we return an error.
		log.Debug(errGetLock, "error", err)
		return reconcile.Result{}, errors.Wrap(resource.IgnoreNotFound(err), errGetLock)
	}

	// NOTE: we add a finalizer and never remove it to guard against
	// a user accidentally deleting the lock, which should never be done.
	// Consider instead recreating the lock if it is not present.
	if err := r.lock.AddFinalizer(ctx, lock); err != nil {
		log.Debug(errAddFinalizer, "error", err)
		return reconcile.Result{RequeueAfter: shortWait}, nil
	}

	log = log.WithValues(
		"uid", lock.GetUID(),
		"version", lock.GetResourceVersion(),
		"name", lock.GetName(),
	)

	dag := r.newDag()
	implied, err := dag.Init(v1.ToNodes(lock.Packages...))
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, errBuildDAG)
	}

	// Make sure we don't have any cyclical imports. If we do, refuse to install
	// additional packages.
	_, err = dag.Sort()
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, errSortDAG)
	}

	if len(implied) == 0 {
		return reconcile.Result{}, nil
	}

	// If we are missing a node, we want to create it. The resolver never
	// modifies the Lock. We only create the first implied node as we will be
	// requeued when it adds itself to the Lock, at which point we will check
	// for missing nodes again.
	dep, ok := implied[0].(*pkgmetav1.Dependency)
	if !ok {
		log.Debug(errInvalidDependency, "error", errors.Errorf(errMissingDependencyFmt, dep.Identifier()))
		return reconcile.Result{}, nil
	}
	c, err := semver.NewConstraint(dep.Constraints)
	if err != nil {
		log.Debug(errInvalidConstraint, "error", err)
		return reconcile.Result{}, nil
	}
	ref, err := name.ParseReference(dep.Package)
	if err != nil {
		log.Debug(errInvalidDependency, "error", err)
		return reconcile.Result{}, nil
	}

	// NOTE: we will be unable to fetch tags for private
	// dependencies because we do not attach any secrets. Consider copying
	// secrets from parent dependencies.
	tags, err := r.fetcher.Tags(ctx, ref)
	if err != nil {
		log.Debug(errFetchTags, "error", err)
		return reconcile.Result{RequeueAfter: shortWait}, nil
	}

	vs := []*semver.Version{}
	for _, r := range tags {
		v, err := semver.NewVersion(r)
		if err != nil {
			// We skip any tags that are not valid semantic versions.
			continue
		}
		vs = append(vs, v)
	}

	sort.Sort(semver.Collection(vs))
	var addVer string
	for _, v := range vs {
		if c.Check(v) {
			addVer = v.Original()
		}
	}

	// NOTE: consider creating event on package revision
	// dictating constraints.
	if addVer == "" {
		log.Debug(errNoValidVersion, errors.Errorf(errNoValidVersionFmt, dep.Identifier(), dep.Constraints))
		return reconcile.Result{}, nil
	}

	var pack v1.Package
	switch dep.Type {
	case pkgmetav1.ProviderPackageType:
		pack = &v1.Provider{}
	default:
		log.Debug(errInvalidPackageType)
		return reconcile.Result{}, nil
	}

	// NOTE: packages are currently created with default
	// settings. This means that a dependency must be publicly available as
	// no packagePullSecrets are set. Settings can be modified manually
	// after dependency creation to address this.
	pack.SetName(nddpkg.ToDNSLabel(ref.Context().RepositoryStr()))
	pack.SetSource(fmt.Sprintf(packageTagFmt, ref.String(), addVer))

	// NOTE: consider making the lock the controller of packages
	// it creates.
	if err := r.client.Create(ctx, pack); err != nil {
		log.Debug(errCreateDependency, "error", err)
		return reconcile.Result{RequeueAfter: shortWait}, nil
	}

	return reconcile.Result{}, nil
}
