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

package manager

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/yndd/ndd-runtime/pkg/meta"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	pkgv1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/yndd/ndd-core/internal/nddpkg"
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

func pullBasedRequeue(p *corev1.PullPolicy) reconcile.Result {
	r := reconcile.Result{}
	if p != nil && *p == corev1.PullAlways {
		r.RequeueAfter = pullWait
	}
	return r
}

const (
	errGetPackage           = "cannot get package"
	errListRevisions        = "cannot list revisions for package"
	errUnpack               = "cannot unpack package"
	errApplyPackageRevision = "cannot apply package revision"
	errGCPackageRevision    = "cannot garbage collect old package revision"

	errUpdateStatus                  = "cannot update package status"
	errUpdateInactivePackageRevision = "cannot update inactive package revision"

	errUnhealthyPackageRevision     = "current package revision is unhealthy"
	errUnknownPackageRevisionHealth = "current package revision health is unknown"
)

// Event reasons.
const (
	reasonList               event.Reason = "ListRevision"
	reasonUnpack             event.Reason = "UnpackPackage"
	reasonTransitionRevision event.Reason = "TransitionRevision"
	reasonGarbageCollect     event.Reason = "GarbageCollect"
	reasonInstall            event.Reason = "InstallPackageRevision"
)

// ReconcilerOption is used to configure the Reconciler.
type ReconcilerOption func(*Reconciler)

// WithNewPackageFn determines the type of package being reconciled.
func WithNewPackageFn(f func() pkgv1.Package) ReconcilerOption {
	return func(r *Reconciler) {
		r.newPackage = f
	}
}

// WithNewPackageRevisionFn determines the type of package being reconciled.
func WithNewPackageRevisionFn(f func() pkgv1.PackageRevision) ReconcilerOption {
	return func(r *Reconciler) {
		r.newPackageRevision = f
	}
}

// WithNewPackageRevisionListFn determines the type of package being reconciled.
func WithNewPackageRevisionListFn(f func() pkgv1.PackageRevisionList) ReconcilerOption {
	return func(r *Reconciler) {
		r.newPackageRevisionList = f
	}
}

// WithRevisioner specifies how the Reconciler should acquire a package image's
// revision name.
func WithRevisioner(d Revisioner) ReconcilerOption {
	return func(r *Reconciler) {
		r.pkg = d
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

// Reconciler reconciles packages.
type Reconciler struct {
	client resource.ClientApplicator
	pkg    Revisioner
	log    logging.Logger
	record event.Recorder

	newPackage             func() pkgv1.Package
	newPackageRevision     func() pkgv1.PackageRevision
	newPackageRevisionList func() pkgv1.PackageRevisionList
}

// Setup adds a controller that reconciles Providers.
func Setup(mgr ctrl.Manager, l logging.Logger, namespace string) error {
	name := "packages/" + strings.ToLower(pkgv1.ProviderGroupKind)
	np := func() pkgv1.Package { return &pkgv1.Provider{} }
	nr := func() pkgv1.PackageRevision { return &pkgv1.ProviderRevision{} }
	nrl := func() pkgv1.PackageRevisionList { return &pkgv1.ProviderRevisionList{} }

	clientset, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		return errors.Wrap(err, "failed to initialize clientset")
	}

	r := NewReconciler(mgr,
		WithNewPackageFn(np),
		WithNewPackageRevisionFn(nr),
		WithNewPackageRevisionListFn(nrl),
		WithRevisioner(NewPackageRevisioner(nddpkg.NewK8sFetcher(clientset, namespace))),
		WithLogger(l.WithValues("controller", name)),
		WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&pkgv1.Provider{}).
		Owns(&pkgv1.ProviderRevision{}).
		Complete(r)
}

// NewReconciler creates a new package reconciler.
func NewReconciler(mgr ctrl.Manager, opts ...ReconcilerOption) *Reconciler {
	r := &Reconciler{
		client: resource.ClientApplicator{
			Client:     mgr.GetClient(),
			Applicator: resource.NewAPIPatchingApplicator(mgr.GetClient()),
		},
		pkg:    NewNopRevisioner(),
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
// +kubebuilder:rbac:groups=pkg.ndd.yndd.io,resources=providerrevisions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pkg.ndd.yndd.io,resources=providerrevisions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=pkg.ndd.yndd.io,resources=providerrevisions/finalizers,verbs=update
// +kubebuilder:rbac:groups=pkg.ndd.yndd.io,resources=providers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pkg.ndd.yndd.io,resources=providers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=pkg.ndd.yndd.io,resources=providers/finalizers,verbs=update

// Reconcile package.
func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) { // nolint:gocyclo
	log := r.log.WithValues("request", req)
	log.Debug("Reconciling", "NameSpace", req.NamespacedName)

	ctx, cancel := context.WithTimeout(ctx, reconcileTimeout)
	defer cancel()

	p := r.newPackage()
	if err := r.client.Get(ctx, req.NamespacedName, p); err != nil {
		// There's no need to requeue if we no longer exist. Otherwise we'll be
		// requeued implicitly because we return an error.
		log.Debug(errGetPackage, "error", err)
		return reconcile.Result{}, errors.Wrap(resource.IgnoreNotFound(err), errGetPackage)
	}

	log = log.WithValues(
		"uid", p.GetUID(),
		"version", p.GetResourceVersion(),
		"name", p.GetName(),
	)

	// Get existing package revisions.
	prs := r.newPackageRevisionList()
	if err := r.client.List(ctx, prs, client.MatchingLabels(map[string]string{pkgv1.ParentLabelKey: p.GetName()})); resource.IgnoreNotFound(err) != nil {
		log.Debug(errListRevisions, "error", err)
		r.record.Event(p, event.Warning(reasonList, errors.Wrap(err, errListRevisions)))
		return reconcile.Result{RequeueAfter: shortWait}, nil
	}

	for _, pr := range prs.GetRevisions() {
		log.Debug("package revisions", "pr", pr)
	}

	// fetch the package from the container registry
	revisionName, err := r.pkg.Revision(ctx, log, p)
	if err != nil {
		p.SetConditions(pkgv1.Unpacking())
		log.Debug(errUnpack, "error", err)
		r.record.Event(p, event.Warning(reasonUnpack, errors.Wrap(err, errUnpack)))
		return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, p), errUpdateStatus)
	}

	log.Debug("revisionName", "revisionName", revisionName)

	if revisionName == "" {
		p.SetConditions(pkgv1.Unpacking())
		r.record.Event(p, event.Normal(reasonUnpack, "Waiting for unpack to complete"))
		return reconcile.Result{RequeueAfter: veryShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), errUpdateStatus)
	}

	// Set the current revision and identifier.
	p.SetCurrentRevision(revisionName)
	p.SetCurrentIdentifier(p.GetSource())

	pr := r.newPackageRevision()
	maxRevision := int64(0)
	oldestRevision := int64(math.MaxInt64)
	oldestRevisionIndex := -1
	revisions := prs.GetRevisions()

	// Check to see if revision already exists.
	for index, rev := range revisions {
		revisionNum := rev.GetRevision()

		// Set max revision to the highest numbered existing revision.
		if revisionNum > maxRevision {
			maxRevision = revisionNum
		}

		// Set oldest revision to the lowest numbered revision and record its
		// index.
		if revisionNum < oldestRevision {
			oldestRevision = revisionNum
			oldestRevisionIndex = index
		}
		// If revision name is same as current revision, then revision already exists.
		if rev.GetName() == p.GetCurrentRevision() {
			pr = rev
			// Finish iterating through all revisions to make sure all
			// non-current revisions are inactive.
			continue
		}
		if rev.GetDesiredState() == pkgv1.PackageRevisionActive {
			// If revision is not the current revision, set to inactive. This
			// should always be done, regardless of the package's revision
			// activation policy.
			rev.SetDesiredState(pkgv1.PackageRevisionInactive)
			if err := r.client.Apply(ctx, rev, resource.MustBeControllableBy(p.GetUID())); err != nil {
				log.Debug(errUpdateInactivePackageRevision, "error", err)
				r.record.Event(p, event.Warning(reasonTransitionRevision, errors.Wrap(err, errUpdateInactivePackageRevision)))
				return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, p), errUpdateStatus)
			}
		}
	}

	// The current revision should always be the highest numbered revision.
	if pr.GetRevision() < maxRevision || maxRevision == 0 {
		pr.SetRevision(maxRevision + 1)
	}

	// Check to see if there are revisions eligible for garbage collection.
	if p.GetRevisionHistoryLimit() != nil &&
		*p.GetRevisionHistoryLimit() != 0 &&
		len(revisions) > (int(*p.GetRevisionHistoryLimit())+1) {
		gcRev := revisions[oldestRevisionIndex]
		// Find the oldest revision and delete it.
		if err := r.client.Delete(ctx, gcRev); err != nil {
			log.Debug(errGCPackageRevision, "error", err)
			r.record.Event(p, event.Warning(reasonGarbageCollect, errors.Wrap(err, errGCPackageRevision)))
			return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, p), errUpdateStatus)
		}
	}

	if pr.GetCondition(pkgv1.ConditionKindPackageHealthy).Status == corev1.ConditionTrue {
		p.SetConditions(pkgv1.Healthy())
		r.record.Event(p, event.Normal(reasonInstall, "Successfully installed package revision"))
	}
	if pr.GetCondition(pkgv1.ConditionKindPackageHealthy).Status == corev1.ConditionFalse {
		p.SetConditions(pkgv1.Unhealthy())
		r.record.Event(p, event.Warning(reasonInstall, errors.New(errUnhealthyPackageRevision)))
	}
	if pr.GetCondition(pkgv1.ConditionKindPackageHealthy).Status == corev1.ConditionUnknown {
		p.SetConditions(pkgv1.UnknownHealth())
		r.record.Event(p, event.Warning(reasonInstall, errors.New(errUnknownPackageRevisionHealth)))
	}

	// Create the non-existent package revision.
	pr.SetName(revisionName)
	labels := map[string]string{pkgv1.ParentLabelKey: p.GetName()}
	if v, ok := p.GetLabels()[pkgv1.CompositeProviderNameLabelKey]; ok {
		labels[pkgv1.CompositeProviderNameLabelKey] = v
	}
	if v, ok := p.GetLabels()[pkgv1.CompositeProviderNamespceLabelKey]; ok {
		labels[pkgv1.CompositeProviderNamespceLabelKey] = v
	}
	pr.SetLabels(labels)
	pr.SetRevisionKind(p.GetKind())
	pr.SetSource(p.GetSource())
	pr.SetPackagePullPolicy(p.GetPackagePullPolicy())
	pr.SetPackagePullSecrets(p.GetPackagePullSecrets())
	pr.SetSkipDependencyResolution(p.GetSkipDependencyResolution())
	pr.SetControllerRef(p.GetControllerRef())

	log.Debug("manager state", "state", pr.GetDesiredState(), "activation policy", p.GetActivationPolicy())
	// If current revision is not active and we have an automatic or undefined
	// activation policy, always activate.
	if pr.GetDesiredState() != pkgv1.PackageRevisionActive && (p.GetActivationPolicy() == nil || *p.GetActivationPolicy() == pkgv1.AutomaticActivation) {
		pr.SetDesiredState(pkgv1.PackageRevisionActive)
	}

	controlRef := meta.AsController(meta.TypedReferenceTo(p, p.GetObjectKind().GroupVersionKind()))
	controlRef.BlockOwnerDeletion = pointer.BoolPtr(true)
	meta.AddOwnerReference(pr, controlRef)
	// create providerrevision with ownerReferences
	if err := r.client.Apply(ctx, pr, resource.MustBeControllableBy(p.GetUID())); err != nil {
		log.Debug(errApplyPackageRevision, "error", err)
		r.record.Event(p, event.Warning(reasonInstall, errors.Wrap(err, errApplyPackageRevision)))
		return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, p), errUpdateStatus)
	}

	p.SetConditions(pkgv1.Active())

	// If current revision is still not active, the package is inactive.
	if pr.GetDesiredState() != pkgv1.PackageRevisionActive {
		p.SetConditions(pkgv1.Inactive())
	}

	// NOTE: when the first package revision is created for a
	// package, the health of the package is not set until the revision reports
	// its health. If updating from an existing revision, the package health
	// will match the health of the old revision until the next reconcile.
	return pullBasedRequeue(p.GetPackagePullPolicy()), errors.Wrap(r.client.Status().Update(ctx, p), errUpdateStatus)
}
