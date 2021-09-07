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

package nn

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	ndddvrv1 "github.com/yndd/ndd-core/apis/dvr/v1"
	"github.com/yndd/ndd-runtime/pkg/event"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-runtime/pkg/meta"
	"github.com/yndd/ndd-runtime/pkg/resource"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	// Finalizer
	finalizer = "networknode.dvr.ndd.yndd.io"

	// default
	defaultGrpcPort = 9999

	// Timers
	reconcileTimeout = 1 * time.Minute
	shortWait        = 30 * time.Second
	veryShortWait    = 5 * time.Second
	longWait         = 1 * time.Minute

	// Errors
	errGetNetworkNode = "cannot get network node resource"
	errUpdateStatus   = "cannot update network node status"

	errAddFinalizer    = "cannot add network node finalizer"
	errRemoveFinalizer = "cannot remove network node finalizer"

	errCredentials = "invalid credentials"

	errDeleteObjects = "cannot delete configmap, servide or deployment"
	errCreateObjects = "cannot create configmap, servide or deployment"

	// Event reasons
	reasonSync event.Reason = "SyncNetworkNode"
)

// ReconcilerOption is used to configure the Reconciler.
type ReconcilerOption func(*Reconciler)

// WithNewNetworkNodeFn determines the type of network node being reconciled.
func WithNewNetworkNodeFn(f func() ndddvrv1.Nn) ReconcilerOption {
	return func(r *Reconciler) {
		r.newNetworkNode = f
	}
}

// WithHooks specifies how the Reconciler should deploy a device driver
func WithHooks(h Hooks) ReconcilerOption {
	return func(r *Reconciler) {
		r.hooks = h
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

// WithValidator specifies how the Reconciler should perform object
// validation.
func WithValidator(v Validator) ReconcilerOption {
	return func(r *Reconciler) {
		r.validator = v
	}
}

// Reconciler reconciles packages.
type Reconciler struct {
	client      client.Client
	nnFinalizer resource.Finalizer
	hooks       Hooks
	validator   Validator
	log         logging.Logger
	record      event.Recorder

	newNetworkNode func() ndddvrv1.Nn
}

// Setup adds a controller that reconciles the Lock.
func Setup(mgr ctrl.Manager, l logging.Logger, namespace string) error {
	name := "dvr/" + strings.ToLower(ndddvrv1.NetworkNodeKind)
	nn := func() ndddvrv1.Nn { return &ndddvrv1.NetworkNode{} }

	r := NewReconciler(mgr,
		WithLogger(l.WithValues("controller", name)),
		WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		WithHooks(NewDeviceDriverHooks(resource.ClientApplicator{
			Client:     mgr.GetClient(),
			Applicator: resource.NewAPIPatchingApplicator(mgr.GetClient()),
		}, l, namespace)),
		WithNewNetworkNodeFn(nn),
		WithValidator(NewNnValidator(resource.ClientApplicator{
			Client:     mgr.GetClient(),
			Applicator: resource.NewAPIPatchingApplicator(mgr.GetClient()),
		}, l)),
	)

	h := &EnqueueRequestForAllDeviceDriversWithRequests{
		client: mgr.GetClient()}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&ndddvrv1.NetworkNode{}).
		Watches(&source.Kind{Type: &ndddvrv1.DeviceDriver{}}, h).
		WithEventFilter(resource.IgnoreUpdateWithoutGenerationChangePredicate()).
		Complete(r)
}

// NewReconciler creates a new package revision reconciler.
func NewReconciler(mgr manager.Manager, opts ...ReconcilerOption) *Reconciler {
	r := &Reconciler{
		client:      mgr.GetClient(),
		nnFinalizer: resource.NewAPIFinalizer(mgr.GetClient(), finalizer),
		log:         logging.NewNopLogger(),
		record:      event.NewNopRecorder(),
	}

	for _, f := range opts {
		f(r)
	}

	return r
}

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;patch;create;update;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;patch;create;update;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;patch;create;update;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=list;watch;get;patch;create;update;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=list;watch;get
// +kubebuilder:rbac:groups="",resources=events,verbs=list;watch;get;patch;create;update;delete
// +kubebuilder:rbac:groups=dvr.ndd.yndd.io,resources=devicedrivers,verbs=get;list;watch
// +kubebuilder:rbac:groups=dvr.ndd.yndd.io,resources=networknodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dvr.ndd.yndd.io,resources=networknodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=dvr.ndd.yndd.io,resources=networknodes/finalizers,verbs=update

// Reconcile network node.
func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) { // nolint:gocyclo
	log := r.log.WithValues("request", req)
	log.Debug("Network Node", "NameSpace", req.NamespacedName)

	nn := &ndddvrv1.NetworkNode{}
	if err := r.client.Get(ctx, req.NamespacedName, nn); err != nil {
		// There's no need to requeue if we no longer exist. Otherwise we'll be
		// requeued implicitly because we return an error.
		log.Debug(errGetNetworkNode, "error", err)
		return reconcile.Result{}, errors.Wrap(resource.IgnoreNotFound(err), errGetNetworkNode)
	}
	log.Debug("Health status", "status", nn.GetCondition(ndddvrv1.ConditionKindDeviceDriverHealthy).Status)

	if meta.WasDeleted(nn) {
		// the k8s garbage collector will delete all the objects that has the ownerreference set
		// as such we dont have to delete the child objects: configmap, service, deployment, serviceaccount, clusterrolebinding

		// Delete finalizer after the object is deleted
		if err := r.nnFinalizer.RemoveFinalizer(ctx, nn); err != nil {
			log.Debug(errRemoveFinalizer, "error", err)
			r.record.Event(nn, event.Warning(reasonSync, errors.Wrap(err, errRemoveFinalizer)))
			return reconcile.Result{RequeueAfter: shortWait}, nil
		}
		return reconcile.Result{Requeue: false}, nil
	}

	// Add a finalizer to newly created objects and update the conditions
	if err := r.nnFinalizer.AddFinalizer(ctx, nn); err != nil {
		log.Debug(errAddFinalizer, "error", err)
		r.record.Event(nn, event.Warning(reasonSync, errors.Wrap(err, errAddFinalizer)))
		return reconcile.Result{RequeueAfter: shortWait}, nil
	}

	// Retrieve the Login details from the network node spec and validate
	// the network node details and build the credentials for communicating
	// to the network node.
	creds, err := r.validator.ValidateCredentials(ctx, nn.Namespace, *nn.Spec.Target.CredentialsName, *nn.Spec.Target.Address)
	log.Debug("Network node creds", "creds", creds, "err", err)
	if err != nil || creds == nil {
		// remove delete the configmap, service, deployment when the service was healthy
		if nn.GetCondition(ndddvrv1.ConditionKindDeviceDriverHealthy).Status == corev1.ConditionTrue {
			if err := r.hooks.Destroy(ctx, nn, &corev1.Container{}); err != nil {
				log.Debug(errDeleteObjects, "error", err)
				r.record.Event(nn, event.Warning(reasonSync, errors.Wrap(err, errDeleteObjects)))
				nn.SetConditions(ndddvrv1.Unhealthy(), ndddvrv1.NotConfigured(), ndddvrv1.NotDiscovered())
				return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, nn), errUpdateStatus)
			}
		}
		log.Debug(errCredentials, "error", err)
		r.record.Event(nn, event.Warning(reasonSync, errors.Wrap(err, errCredentials)))
		nn.SetConditions(ndddvrv1.Unhealthy(), ndddvrv1.NotConfigured(), ndddvrv1.NotDiscovered())
		return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, nn), errUpdateStatus)
	}

	// validate device driver information
	// NOTE: the parameters are required in the api and will get defaults if not specified, so we dont have to add validation
	// if they exist in the api or not
	c, err := r.validator.ValidateDeviceDriver(ctx, nn.Namespace, nn.Name, string(*nn.Spec.DeviceDriverKind), *nn.Spec.GrpcServerPort)
	log.Debug("Validate device driver", "containerInfo", c, "err", err)
	if err != nil || c == nil {
		if nn.GetCondition(ndddvrv1.ConditionKindDeviceDriverHealthy).Status == corev1.ConditionTrue {
			if err := r.hooks.Destroy(ctx, nn, &corev1.Container{}); err != nil {
				log.Debug(errDeleteObjects, "error", err)
				r.record.Event(nn, event.Warning(reasonSync, errors.Wrap(err, errDeleteObjects)))
				nn.SetConditions(ndddvrv1.Unhealthy(), ndddvrv1.NotConfigured(), ndddvrv1.NotDiscovered())
				return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, nn), errUpdateStatus)
			}
		}
	}

	if nn.GetCondition(ndddvrv1.ConditionKindDeviceDriverHealthy).Status == corev1.ConditionTrue &&
		nn.GetCondition(ndddvrv1.ConditionKindDeviceDriverConfigured).Status == corev1.ConditionTrue {
		// this is most likely a restart of the ndd-core, given the status of the deployment of the network
		// node is ok we should stop the reconciliation here
		return reconcile.Result{}, nil
	}

	// when everything is validated we want to bring the deployment in healthy status by all means
	if err := r.hooks.Deploy(ctx, nn, c); err != nil {
		log.Debug(errCreateObjects, "error", err)
		r.record.Event(nn, event.Warning(reasonSync, errors.Wrap(err, errCreateObjects)))
		nn.SetConditions(ndddvrv1.Unhealthy(), ndddvrv1.NotConfigured(), ndddvrv1.NotDiscovered())
		return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, nn), errUpdateStatus)
	}
	r.record.Event(nn, event.Normal(reasonSync, "Successfully deployed network device driver"))
	nn.SetConditions(ndddvrv1.Healthy(), ndddvrv1.NotConfigured(), ndddvrv1.NotDiscovered())
	return reconcile.Result{}, errors.Wrap(r.client.Status().Update(ctx, nn), errUpdateStatus)
}
