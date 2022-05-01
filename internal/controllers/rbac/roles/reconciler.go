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

package roles

import (
	"context"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	pkgv1 "github.com/yndd/ndd-core/apis/pkg/v1"
	v1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/yndd/ndd-runtime/pkg/event"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-runtime/pkg/meta"
	"github.com/yndd/ndd-runtime/pkg/resource"
	rbacv1 "k8s.io/api/rbac/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	kcontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	// timers
	shortWait = 30 * time.Second

	timeout        = 2 * time.Minute
	maxConcurrency = 5

	//errprs
	errGetPR               = "cannot get Package Revision"
	errListCRDs            = "cannot list CustomResourceDefinitions"
	errGetPkgMeta          = "cannot get package meta cr"
	errApplyRole           = "cannot apply ClusterRole"
	errValidatePermissions = "cannot validate permission requests"
	errRejectedPermission  = "refusing to apply any RBAC roles due to request for disallowed permission"

	// evnts reason
	reasonApplyRoles event.Reason = "ApplyClusterRoles"
)

// A PermissionRequestsValidator validates requested RBAC rules.
type PermissionRequestsValidator interface {
	// ValidatePermissionRequests validates the supplied slice of RBAC rules. It
	// returns a slice of any rejected (i.e. disallowed) rules. It returns an
	// error if it is unable to validate permission requests.
	ValidatePermissionRequests(ctx context.Context, requested ...rbacv1.PolicyRule) ([]Rule, error)
}

// A PermissionRequestsValidatorFn validates requested RBAC rules.
type PermissionRequestsValidatorFn func(ctx context.Context, requested ...rbacv1.PolicyRule) ([]Rule, error)

// ValidatePermissionRequests validates the supplied slice of RBAC rules. It
// returns a slice of any rejected (i.e. disallowed) rules. It returns an error
// if it is unable to validate permission requests.
func (fn PermissionRequestsValidatorFn) ValidatePermissionRequests(ctx context.Context, requested ...rbacv1.PolicyRule) ([]Rule, error) {
	return fn(ctx, requested...)
}

// A ClusterRoleRenderer renders ClusterRoles for the given CRDs.
type ClusterRoleRenderer interface {
	// RenderClusterRoles for the supplied CRDs.
	RenderClusterRoles(pr *v1.PackageRevision, crds []extv1.CustomResourceDefinition) []rbacv1.ClusterRole
}

// A ClusterRoleRenderFn renders ClusterRoles for the supplied CRDs.
type ClusterRoleRenderFn func(pr *v1.PackageRevision, crds []extv1.CustomResourceDefinition) []rbacv1.ClusterRole

// RenderClusterRoles renders ClusterRoles for the supplied CRDs.
func (fn ClusterRoleRenderFn) RenderClusterRoles(pr *v1.PackageRevision, crds []extv1.CustomResourceDefinition) []rbacv1.ClusterRole {
	return fn(pr, crds)
}

// Setup adds a controller that reconciles a ProviderRevision by creating a
// series of opinionated ClusterRoles that may be bound to allow access to the
// resources it defines.
func SetupProvider(mgr ctrl.Manager, log logging.Logger, allowClusterRole string) error {
	name := "rbac/" + strings.ToLower(v1.ProviderRevisionGroupKind)
	np := func() v1.Package { return &v1.Provider{} }
	nr := func() v1.PackageRevision { return &v1.ProviderRevision{} }
	nrl := func() v1.PackageRevisionList { return &v1.ProviderRevisionList{} }

	if allowClusterRole != "" {

		h := &EnqueueRequestForAllRevisionsWithRequests{
			client:          mgr.GetClient(),
			clusterRoleName: allowClusterRole}

		return ctrl.NewControllerManagedBy(mgr).
			Named(name).
			For(&v1.ProviderRevision{}).
			Owns(&rbacv1.ClusterRole{}).
			Watches(&source.Kind{Type: &rbacv1.ClusterRole{}}, h).
			WithOptions(kcontroller.Options{MaxConcurrentReconciles: maxConcurrency}).
			Complete(NewReconciler(mgr,
				WithNewPackageFn(np),
				WithNewPackageRevisionFn(nr),
				WithNewPackageRevisionListFn(nrl),
				WithLogger(log.WithValues("controller", name)),
				WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
				WithPermissionRequestsValidator(NewClusterRoleBackedValidator(mgr.GetClient(), allowClusterRole)),
			))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1.ProviderRevision{}).
		Owns(&rbacv1.ClusterRole{}).
		WithOptions(kcontroller.Options{MaxConcurrentReconciles: maxConcurrency}).
		Complete(NewReconciler(mgr,
			WithNewPackageFn(np),
			WithNewPackageRevisionFn(nr),
			WithNewPackageRevisionListFn(nrl),
			WithLogger(log.WithValues("controller", name)),
			WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

// Setup adds a controller that reconciles a IntentRevision by creating a
// series of opinionated ClusterRoles that may be bound to allow access to the
// resources it defines.
func SetupIntent(mgr ctrl.Manager, log logging.Logger, allowClusterRole string) error {
	name := "rbac/" + strings.ToLower(v1.IntentRevisionGroupKind)
	np := func() v1.Package { return &v1.Intent{} }
	nr := func() v1.PackageRevision { return &v1.IntentRevision{} }
	nrl := func() v1.PackageRevisionList { return &v1.IntentRevisionList{} }

	if allowClusterRole != "" {

		h := &EnqueueRequestForAllRevisionsWithRequests{
			client:          mgr.GetClient(),
			clusterRoleName: allowClusterRole}

		return ctrl.NewControllerManagedBy(mgr).
			Named(name).
			For(&v1.IntentRevision{}).
			Owns(&rbacv1.ClusterRole{}).
			Watches(&source.Kind{Type: &rbacv1.ClusterRole{}}, h).
			WithOptions(kcontroller.Options{MaxConcurrentReconciles: maxConcurrency}).
			Complete(NewReconciler(mgr,
				WithNewPackageFn(np),
				WithNewPackageRevisionFn(nr),
				WithNewPackageRevisionListFn(nrl),
				WithLogger(log.WithValues("controller", name)),
				WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
				WithPermissionRequestsValidator(NewClusterRoleBackedValidator(mgr.GetClient(), allowClusterRole)),
			))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1.IntentRevision{}).
		Owns(&rbacv1.ClusterRole{}).
		WithOptions(kcontroller.Options{MaxConcurrentReconciles: maxConcurrency}).
		Complete(NewReconciler(mgr,
			WithNewPackageFn(np),
			WithNewPackageRevisionFn(nr),
			WithNewPackageRevisionListFn(nrl),
			WithLogger(log.WithValues("controller", name)),
			WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

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

// WithClusterRoleRenderer specifies how the Reconciler should render RBAC
// ClusterRoles.
func WithClusterRoleRenderer(rr ClusterRoleRenderer) ReconcilerOption {
	return func(r *Reconciler) {
		r.rbac.ClusterRoleRenderer = rr
	}
}

// WithPermissionRequestsValidator specifies how the Reconciler should validate
// requests for extra RBAC permissions.
func WithPermissionRequestsValidator(rv PermissionRequestsValidator) ReconcilerOption {
	return func(r *Reconciler) {
		r.rbac.PermissionRequestsValidator = rv
	}
}

// WithNewPackageFn determines the type of package being reconciled.
func WithNewPackageFn(f func() v1.Package) ReconcilerOption {
	return func(r *Reconciler) {
		r.newPackage = f
	}
}

// WithNewPackageRevisionFn determines the type of package being reconciled.
func WithNewPackageRevisionFn(f func() v1.PackageRevision) ReconcilerOption {
	return func(r *Reconciler) {
		r.newPackageRevision = f
	}
}

// WithNewPackageRevisionListFn determines the type of package being reconciled.
func WithNewPackageRevisionListFn(f func() v1.PackageRevisionList) ReconcilerOption {
	return func(r *Reconciler) {
		r.newPackageRevisionList = f
	}
}

// NewReconciler returns a Reconciler of PackageRevisions.
func NewReconciler(mgr manager.Manager, opts ...ReconcilerOption) *Reconciler {
	r := &Reconciler{
		// TODO(negz): Is Updating appropriate here? Probably.
		client: resource.ClientApplicator{
			Client:     mgr.GetClient(),
			Applicator: resource.NewAPIUpdatingApplicator(mgr.GetClient()),
		},

		rbac: rbac{
			PermissionRequestsValidator: PermissionRequestsValidatorFn(VerySecureValidator),
			ClusterRoleRenderer:         ClusterRoleRenderFn(RenderClusterRoles),
		},

		log:    logging.NewNopLogger(),
		record: event.NewNopRecorder(),
	}

	for _, f := range opts {
		f(r)
	}
	return r
}

type rbac struct {
	PermissionRequestsValidator
	ClusterRoleRenderer
}

// A Reconciler reconciles PackageRevisions.
type Reconciler struct {
	client resource.ClientApplicator
	rbac   rbac

	log    logging.Logger
	record event.Recorder

	newPackage             func() v1.Package
	newPackageRevision     func() v1.PackageRevision
	newPackageRevisionList func() v1.PackageRevisionList
}

// Reconcile a PackageRevision by creating a series of opinionated ClusterRoles
// that may be bound to allow access to the resources it defines.
func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) { //nolint:gocyclo
	log := r.log.WithValues("request", req)
	log.Debug("Reconciling")

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	pr := r.newPackageRevision()

	//pr := &v1.ProviderRevision{}
	if err := r.client.Get(ctx, req.NamespacedName, pr); err != nil {
		// In case object is not found, most likely the object was deleted and
		// then disappeared while the event was in the processing queue. We
		// don't need to take any action in that case.
		log.Debug(errGetPR, "error", err)
		return reconcile.Result{}, errors.Wrap(resource.IgnoreNotFound(err), errGetPR)
	}

	log = log.WithValues(
		"uid", pr.GetUID(),
		"version", pr.GetResourceVersion(),
		"name", pr.GetName(),
		"owner", pr.GetOwnerReferences(),
	)

	if meta.WasDeleted(pr) {
		// There's nothing to do if our PR is being deleted. Any ClusterRoles
		// we created will be garbage collected by Kubernetes.
		return reconcile.Result{Requeue: false}, nil
	}

	// get the controller config to operate
	/*
		if pr.GetKind() != pkgv1.IntentRevisionKind && len(pr.GetOwnerReferences()) != 0 {
			log.Debug("provider kind")
			pkgMeta := &pkgmetav1.Provider{}
			if err := r.client.Get(ctx, types.NamespacedName{Namespace: os.Getenv("POD_NAMESPACE"), Name: pr.GetOwnerReferences()[0].Name}, pkgMeta); err != nil {
				log.Debug(errGetPkgMeta, "error", err)
				return reconcile.Result{Requeue: true}, errors.Wrap(err, errGetPkgMeta)
			}
			log.Debug("controller config", "config", pkgMeta.Spec)
		}
	*/

	l := &extv1.CustomResourceDefinitionList{}
	if err := r.client.List(ctx, l); err != nil {
		log.Debug(errListCRDs, "error", err)
		r.record.Event(pr, event.Warning(reasonApplyRoles, errors.Wrap(err, errListCRDs)))
		return reconcile.Result{RequeueAfter: shortWait}, nil
	}

	// Filter down to the CRDs that are owned by this PackageRevision - i.e.
	// those that it may become the active revision for.
	crds := make([]extv1.CustomResourceDefinition, 0)
	for _, crd := range l.Items {
		for _, ref := range crd.GetOwnerReferences() {
			if ref.UID == pr.GetUID() {
				crds = append(crds, crd)
			}
		}
	}

	// TODO validate permission request handling
	/*
		rejected, err := r.rbac.ValidatePermissionRequests(ctx, pr.GetPermissionsRequests()...)
		if err != nil {
			log.Debug(errValidatePermissions, "error", err)
			r.record.Event(pr, event.Warning(reasonApplyRoles, errors.Wrap(err, errValidatePermissions)))
			return reconcile.Result{RequeueAfter: shortWait}, nil
		}

		for _, rule := range rejected {
			log.Debug(errRejectedPermission, "rule", rule)
			r.record.Event(pr, event.Warning(reasonApplyRoles, errors.Errorf("%s %s", errRejectedPermission, rule)))
		}

		// We return early and don't grant _any_ RBAC permissions if we would reject
		// any requested permission. It's better for the provider to be completely
		// and obviously broken than for it to be subtly broken in a way that may
		// not surface immediately, i.e. due to missing an RBAC permission it only
		// occasionally needs. There's no need to requeue - the revisions requests
		// won't change, and we're watching the ClusterRole of allowed requests.
		if len(rejected) > 0 {
			return reconcile.Result{Requeue: false}, nil
		}
	*/

	if pr.GetKind() == pkgv1.ProviderRevisionKind {
		log.Debug("permission requests1", "pr", pr.GetPermissionsRequests())
		provRev, _ := pr.(*pkgv1.ProviderRevision)
		log.Debug("permission requests2", "pr", provRev.Status.PermissionRequests)

		for _, prr := range pr.GetPermissionsRequests() {
			if len(prr.Resources) == 0 {
				log.Debug("permission requests3", "pr", provRev.Status.PermissionRequests)
			}
		}
	}

	for _, cr := range r.rbac.RenderClusterRoles(&pr, crds) {
		log.Debug("clusterrole", "cr", cr)
		cr := cr // Pin range variable so we can take its address.
		log = log.WithValues("role-name", cr.GetName())
		err := r.client.Apply(ctx, &cr, resource.MustBeControllableBy(pr.GetUID()), resource.AllowUpdateIf(ClusterRolesDiffer))
		if resource.IsNotAllowed(err) {
			log.Debug("Skipped no-op RBAC ClusterRole apply")
			continue
		}
		if err != nil {
			log.Debug(errApplyRole, "error", err)
			r.record.Event(pr, event.Warning(reasonApplyRoles, errors.Wrap(err, errApplyRole)))
			return reconcile.Result{RequeueAfter: shortWait}, nil
		}
		log.Debug("Applied RBAC ClusterRole")
	}

	// TODO: Add a condition that indicates the RBAC manager is
	// managing cluster roles for this PackageRevision?
	r.record.Event(pr, event.Normal(reasonApplyRoles, "Applied RBAC ClusterRoles"))

	// There's no need to requeue explicitly - we're watching all PRs.
	return reconcile.Result{Requeue: false}, nil
}

// ClusterRolesDiffer returns true if the supplied objects are different
// ClusterRoles. We consider ClusterRoles to be different if their labels and
// rules do not match.
func ClusterRolesDiffer(current, desired runtime.Object) bool {
	c := current.(*rbacv1.ClusterRole)
	d := desired.(*rbacv1.ClusterRole)
	return !cmp.Equal(c.GetLabels(), d.GetLabels()) || !cmp.Equal(c.Rules, d.Rules)
}
