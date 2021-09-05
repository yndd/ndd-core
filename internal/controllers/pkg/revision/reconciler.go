/*
Copyright 2021 Wim Henderickx.

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

package revision

import (
	"context"
	"strings"
	"time"

	"github.com/yndd/ndd-core/internal/dag"
	"github.com/yndd/ndd-core/internal/nddpkg"
	"github.com/yndd/ndd-core/internal/version"
	"github.com/yndd/ndd-runtime/pkg/event"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-runtime/pkg/meta"
	"github.com/yndd/ndd-runtime/pkg/parser"
	"github.com/yndd/ndd-runtime/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	pkgmetav1 "github.com/yndd/ndd-core/apis/pkg/meta/v1"
	v1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	// Finalizer
	finalizer = "revision.pkg.ndd.yndd.io"

	// Timers
	reconcileTimeout = 1 * time.Minute
	shortWait        = 30 * time.Second
	longWait         = 1 * time.Minute

	// Errors
	errGetPackageRevision = "cannot get package revision"
	errUpdateStatus       = "cannot update package revision status"

	errDeleteCache = "cannot remove package image from cache"

	errAddFinalizer    = "cannot add package revision finalizer"
	errRemoveFinalizer = "cannot remove package revision finalizer"

	errInitParserBackend = "cannot initialize parser backend"
	errParsePackage      = "cannot parse package contents"
	errNotOneMeta        = "cannot install package with multiple meta types"

	errPreHook  = "cannot run pre establish hook for package"
	errPostHook = "cannot run post establish hook for package"

	errEstablishControl = "cannot establish control of object"

	// Event reasons
	reasonParse        event.Reason = "ParsePackage"
	reasonLint         event.Reason = "LintPackage"
	reasonDependencies event.Reason = "ResolveDependencies"
	reasonSync         event.Reason = "SyncPackage"
)

// ReconcilerOption is used to configure the Reconciler.
type ReconcilerOption func(*Reconciler)

// WithClientApplicator specifies how the Reconciler should interact with the
// Kubernetes API.
func WithClientApplicator(ca resource.ClientApplicator) ReconcilerOption {
	return func(r *Reconciler) {
		r.client = ca
	}
}

// WithCache specifies how the Reconcile should cache package contents.
func WithCache(c nddpkg.Cache) ReconcilerOption {
	return func(r *Reconciler) {
		r.cache = c
	}
}

// WithNewPackageRevisionFn determines the type of package being reconciled.
func WithNewPackageRevisionFn(f func() v1.PackageRevision) ReconcilerOption {
	return func(r *Reconciler) {
		r.newPackageRevision = f
	}
}

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
		r.revision = f
	}
}

// WithDependencyManager specifies how the Reconciler should manage dependencies.
func WithDependencyManager(m DependencyManager) ReconcilerOption {
	return func(r *Reconciler) {
		r.lock = m
	}
}

// WithHooks specifies how the Reconciler should perform pre and post object
// establishment operations.
func WithHooks(h Hooks) ReconcilerOption {
	return func(r *Reconciler) {
		r.hook = h
	}
}

// WithEstablisher specifies how the Reconciler should establish package resources.
func WithEstablisher(e Establisher) ReconcilerOption {
	return func(r *Reconciler) {
		r.objects = e
	}
}

// WithParser specifies how the Reconciler should parse a package.
func WithParser(p parser.Parser) ReconcilerOption {
	return func(r *Reconciler) {
		r.parser = p
	}
}

// WithParserBackend specifies how the Reconciler should parse a package.
func WithParserBackend(p parser.Backend) ReconcilerOption {
	return func(r *Reconciler) {
		r.backend = p
	}
}

// WithLinter specifies how the Reconciler should lint a package.
func WithLinter(l parser.Linter) ReconcilerOption {
	return func(r *Reconciler) {
		r.linter = l
	}
}

// WithVersioner specifies how the Reconciler should fetch the current
// Ndd version.
func WithVersioner(v version.Operations) ReconcilerOption {
	return func(r *Reconciler) {
		r.versioner = v
	}
}

// Reconciler reconciles packages.
type Reconciler struct {
	client    client.Client
	cache     nddpkg.Cache
	revision  resource.Finalizer
	lock      DependencyManager
	hook      Hooks
	objects   Establisher
	parser    parser.Parser
	linter    parser.Linter
	versioner version.Operations
	backend   parser.Backend
	log       logging.Logger
	record    event.Recorder

	newPackageRevision func() v1.PackageRevision
}

// SetupProviderRevision adds a controller that reconciles ProviderRevisions.
func SetupProviderRevision(mgr ctrl.Manager, l logging.Logger, cache nddpkg.Cache, namespace string) error {
	name := "packages/" + strings.ToLower(v1.ProviderRevisionGroupKind)
	nr := func() v1.PackageRevision { return &v1.ProviderRevision{} }

	clientset, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		return errors.Wrap(err, "failed to initialize host clientset with in cluster config")
	}

	metaScheme, err := nddpkg.BuildMetaScheme()
	if err != nil {
		return errors.New("cannot build meta scheme for package parser")
	}
	objScheme, err := nddpkg.BuildObjectScheme()
	if err != nil {
		return errors.New("cannot build object scheme for package parser")
	}

	r := NewReconciler(mgr,
		WithCache(cache),
		WithDependencyManager(NewPackageDependencyManager(mgr.GetClient(), dag.NewMapDag, pkgmetav1.ProviderPackageType)),
		WithHooks(NewProviderHooks(resource.ClientApplicator{
			Client:     mgr.GetClient(),
			Applicator: resource.NewAPIPatchingApplicator(mgr.GetClient()),
		}, namespace)),
		WithNewPackageRevisionFn(nr),
		WithParser(parser.New(metaScheme, objScheme)),
		WithParserBackend(NewImageBackend(cache, nddpkg.NewK8sFetcher(clientset, namespace))),
		WithLinter(nddpkg.NewProviderLinter()),
		WithLogger(l.WithValues("controller", name)),
		WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1.ProviderRevision{}).
		Complete(r)
}

// NewReconciler creates a new package revision reconciler.
func NewReconciler(mgr manager.Manager, opts ...ReconcilerOption) *Reconciler {

	r := &Reconciler{
		client:    mgr.GetClient(),
		cache:     nddpkg.NewNopCache(),
		revision:  resource.NewAPIFinalizer(mgr.GetClient(), finalizer),
		hook:      NewNopHooks(),
		objects:   NewAPIEstablisher(mgr.GetClient()),
		parser:    parser.New(nil, nil),
		linter:    parser.NewPackageLinter(nil, nil, nil),
		versioner: version.New(),
		log:       logging.NewNopLogger(),
		record:    event.NewNopRecorder(),
	}

	for _, f := range opts {
		f(r)
	}

	return r
}

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pkg.ndd.yndd.io,resources=locks,verbs=get;list;watch;create;update;patch;delete

// Reconcile package revision.
func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) { // nolint:gocyclo
	log := r.log.WithValues("request", req)
	log.Debug("Revision Reconciling Package Revision", "NameSpace", req.NamespacedName)

	ctx, cancel := context.WithTimeout(ctx, reconcileTimeout)
	defer cancel()

	pr := r.newPackageRevision()
	if err := r.client.Get(ctx, req.NamespacedName, pr); err != nil {
		// There's no need to requeue if we no longer exist. Otherwise we'll be
		// requeued implicitly because we return an error.
		log.Debug(errGetPackageRevision, "error", err)
		return reconcile.Result{}, errors.Wrap(resource.IgnoreNotFound(err), errGetPackageRevision)
	}
	log.Debug("Package Revision", "PR", pr)

	if meta.WasDeleted(pr) {
		// NOTE: In the event that a pre-cached package was used for this revision,
		// delete will not remove the pre-cached package image from the cache
		// unless it has the same name as the provider revision. Delete will not
		// return an error so we will remove finalizer and leave the image in
		// the cache.
		if err := r.cache.Delete(pr.GetName()); err != nil {
			log.Debug(errDeleteCache, "error", err)
			r.record.Event(pr, event.Warning(reasonSync, errors.Wrap(err, errDeleteCache)))
			return reconcile.Result{RequeueAfter: shortWait}, nil
		}
		// NOTE: if we were previously marked as inactive, we likely
		// already removed self. If we skipped dependency resolution, we will
		// not be present in the lock.
		if err := r.lock.RemoveSelf(ctx, pr); err != nil {
			pr.SetConditions(v1.Unhealthy())
			r.record.Event(pr, event.Warning(reasonLint, err))
			return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, pr), errUpdateStatus)
		}
		if err := r.revision.RemoveFinalizer(ctx, pr); err != nil {
			log.Debug(errRemoveFinalizer, "error", err)
			r.record.Event(pr, event.Warning(reasonSync, errors.Wrap(err, errRemoveFinalizer)))
			return reconcile.Result{RequeueAfter: shortWait}, nil
		}
		return reconcile.Result{Requeue: false}, nil
	}

	if err := r.revision.AddFinalizer(ctx, pr); err != nil {
		log.Debug(errAddFinalizer, "error", err)
		r.record.Event(pr, event.Warning(reasonSync, errors.Wrap(err, errAddFinalizer)))
		return reconcile.Result{RequeueAfter: shortWait}, nil
	}

	log = log.WithValues(
		"uid", pr.GetUID(),
		"version", pr.GetResourceVersion(),
		"name", pr.GetName(),
	)

	// Initialize parser backend to obtain package contents.
	reader, err := r.backend.Init(ctx, PackageRevision(pr))
	if err != nil {
		log.Debug(errInitParserBackend, "error", err)
		r.record.Event(pr, event.Warning(reasonParse, errors.Wrap(err, errInitParserBackend)))
		// Requeue after shortWait because we may be waiting for parent package
		// controller to recreate Pod.
		pr.SetConditions(v1.Unhealthy())
		return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, pr), errUpdateStatus)
	}

	// Parse package contents.

	pkg, err := r.parser.Parse(ctx, reader)
	if err != nil {
		log.Debug(errParsePackage, "error", err)
		r.record.Event(pr, event.Warning(reasonParse, errors.Wrap(err, errParsePackage)))
		pr.SetConditions(v1.Unhealthy())
		return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, pr), errUpdateStatus)
	}

	// Lint package using package-specific linter.
	if err := r.linter.Lint(pkg); err != nil {
		r.record.Event(pr, event.Warning(reasonLint, err))
		// NOTE: a failed lint typically will require manual
		// intervention, but on the off chance that we read pod logs early,
		// which caused a linting failure, we will requeue after long wait.
		pr.SetConditions(v1.Unhealthy())
		return reconcile.Result{RequeueAfter: longWait}, errors.Wrap(r.client.Status().Update(ctx, pr), errUpdateStatus)
	}

	// NOTE: the linter should check this property already, but if a
	// consumer forgets to pass an option to guarantee one meta object, we check
	// here to avoid a potential panic on 0 index below.
	if len(pkg.GetMeta()) != 1 {
		r.record.Event(pr, event.Warning(reasonLint, errors.New(errNotOneMeta)))
		pr.SetConditions(v1.Unhealthy())
		return reconcile.Result{RequeueAfter: longWait}, errors.Wrap(r.client.Status().Update(ctx, pr), errUpdateStatus)
	}

	pkgMeta, _ := nddpkg.TryConvert(pkg.GetMeta()[0], &pkgmetav1.Provider{})

	// Check status of package dependencies unless package specifies to skip
	// resolution.
	if pr.GetSkipDependencyResolution() != nil && !*pr.GetSkipDependencyResolution() {
		found, installed, invalid, err := r.lock.Resolve(ctx, pkgMeta, pr)
		pr.SetDependencyStatus(int64(found), int64(installed), int64(invalid))
		if err != nil {
			pr.SetConditions(v1.UnknownHealth())
			r.record.Event(pr, event.Warning(reasonDependencies, err))
			return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, pr), errUpdateStatus)
		}
	}

	if err := r.hook.Pre(ctx, pkgMeta, pr); err != nil {
		log.Debug(errPreHook, "error", err)
		r.record.Event(pr, event.Warning(reasonSync, errors.Wrap(err, errPreHook)))
		pr.SetConditions(v1.Unhealthy())
		return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, pr), errUpdateStatus)
	}

	// Establish control or ownership of objects.
	refs, err := r.objects.Establish(ctx, pkg.GetObjects(), pr, pr.GetDesiredState() == v1.PackageRevisionActive)
	if err != nil {
		log.Debug(errEstablishControl, "error", err)
		r.record.Event(pr, event.Warning(reasonSync, errors.Wrap(err, errEstablishControl)))
		pr.SetConditions(v1.Unhealthy())
		return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, pr), errUpdateStatus)
	}

	// Update object list in package revision status with objects for which
	// ownership or control has been established.
	pr.SetObjects(refs)

	if err := r.hook.Post(ctx, pkgMeta, pr); err != nil {
		log.Debug(errPostHook, "error", err)
		r.record.Event(pr, event.Warning(reasonSync, errors.Wrap(err, errPostHook)))
		pr.SetConditions(v1.Unhealthy())
		return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, pr), errUpdateStatus)
	}

	r.record.Event(pr, event.Normal(reasonSync, "package revision successfully configured and healthy"))
	pr.SetConditions(v1.Healthy())
	return reconcile.Result{RequeueAfter: longWait}, errors.Wrap(r.client.Status().Update(ctx, pr), errUpdateStatus)
}
