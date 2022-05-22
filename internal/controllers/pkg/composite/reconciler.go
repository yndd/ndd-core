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

package composite

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	pkgv1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/yndd/ndd-runtime/pkg/event"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-runtime/pkg/resource"
)

const (
	reconcileTimeout = 1 * time.Minute

	shortWait     = 30 * time.Second
	veryShortWait = 5 * time.Second
	pullWait      = 1 * time.Minute
)

const (
	errListProviders        = "cannot list providers"
	errGetCompositeProvider = "cannot get composite provider"
	errUpdateProvider       = "cannot update provider"
	errDeleteProvider       = "cannot delete provider"
	errUpdateStatus         = "cannot update composite provider status"
)

// Event reasons.
const (
	reasonList event.Reason = "ListProvider"
	//reasonUnpack             event.Reason = "UnpackPackage"
	reasonUpdateProvider event.Reason = "UpdateProvider"
	reasonDeleteProvider event.Reason = "DeleteProvider"
	//reasonGarbageCollect     event.Reason = "GarbageCollect"
	//reasonInstall            event.Reason = "InstallPackageRevision"
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

// Reconciler reconciles packages.
type Reconciler struct {
	client resource.ClientApplicator
	log    logging.Logger
	record event.Recorder
}

// SetupProviderComposite adds a controller that reconciles CompositeProviders.
func SetupCompositeProvider(mgr ctrl.Manager, l logging.Logger, namespace string) error {
	name := "packages/" + strings.ToLower(pkgv1.ProviderGroupKind)

	r := NewReconciler(mgr,
		WithLogger(l.WithValues("controller", name)),
		WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&pkgv1.CompositeProvider{}).
		Owns(&pkgv1.CompositeProvider{}).
		Complete(r)
}

// NewReconciler creates a new package reconciler.
func NewReconciler(mgr ctrl.Manager, opts ...ReconcilerOption) *Reconciler {
	r := &Reconciler{
		client: resource.ClientApplicator{
			Client:     mgr.GetClient(),
			Applicator: resource.NewAPIPatchingApplicator(mgr.GetClient()),
		},
		log:    logging.NewNopLogger(),
		record: event.NewNopRecorder(),
	}

	for _, f := range opts {
		f(r)
	}

	return r
}

// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pkg.ndd.yndd.io,resources=compositeproviders,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pkg.ndd.yndd.io,resources=compositeproviders/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=pkg.ndd.yndd.io,resources=compositeproviders/finalizers,verbs=update

// Reconcile package.
func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) { // nolint:gocyclo
	log := r.log.WithValues("request", req)
	log.Debug("Reconciling", "NameSpace", req.NamespacedName)

	ctx, cancel := context.WithTimeout(ctx, reconcileTimeout)
	defer cancel()

	cp := &pkgv1.CompositeProvider{}
	if err := r.client.Get(ctx, req.NamespacedName, cp); err != nil {
		// There's no need to requeue if we no longer exist. Otherwise we'll be
		// requeued implicitly because we return an error.
		log.Debug(errGetCompositeProvider, "error", err)
		return reconcile.Result{}, errors.Wrap(resource.IgnoreNotFound(err), errGetCompositeProvider)
	}

	log = log.WithValues(
		"uid", cp.GetUID(),
		"version", cp.GetResourceVersion(),
		"name", cp.GetName(),
	)

	pl := &pkgv1.ProviderList{}
	if err := r.client.List(ctx, pl, client.MatchingLabels(map[string]string{
		pkgv1.CompositeProviderGroupKind: types.NamespacedName{Namespace: cp.Namespace, Name: cp.Name}.String(),
	})); resource.IgnoreNotFound(err) != nil {
		log.Debug(errListProviders, "error", err)
		r.record.Event(cp, event.Warning(reasonList, errors.Wrap(err, errListProviders)))
		return reconcile.Result{RequeueAfter: shortWait}, nil
	}

	newProviders := []string{}
	for _, pkg := range cp.Spec.Packages {
		p := renderProvider(cp, pkg)
		if err := r.client.Apply(ctx, p, resource.MustBeControllableBy(cp.GetUID())); err != nil {
			log.Debug(errUpdateProvider, "error", err)
			r.record.Event(cp, event.Warning(reasonUpdateProvider, errors.Wrap(err, errUpdateProvider)))
			return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, p), errUpdateStatus)
		}
		newProviders = append(newProviders, getProviderName(cp.Name, pkg.Name))
	}

	// delete the providers if they are no longer required
	for _, activeProvider := range pl.Items {
		found := false
		for _, newProvider := range newProviders {
			if activeProvider.Name == newProvider {
				found = true
				break
			}
		}
		if !found {
			p := renderProvider(cp, activeProvider.Spec.PackageSpec)
			if err := r.client.Delete(ctx, p); err != nil {
				log.Debug(errUpdateProvider, "error", err)
				r.record.Event(cp, event.Warning(reasonDeleteProvider, errors.Wrap(err, errDeleteProvider)))
				return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, p), errDeleteProvider)
			}
		}
	}

	return reconcile.Result{}, errors.Wrap(r.client.Status().Update(ctx, cp), errUpdateStatus)
}
