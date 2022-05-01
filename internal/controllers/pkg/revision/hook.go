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

package revision

import (
	"context"

	"github.com/pkg/errors"
	pkgmetav1 "github.com/yndd/ndd-core/apis/pkg/meta/v1"
	v1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/yndd/ndd-core/internal/nddpkg"
	nddv1 "github.com/yndd/ndd-runtime/apis/common/v1"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-runtime/pkg/resource"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const (
	errNotProvider                   = "not a provider package"
	errNotProviderRevision           = "not a provider revision"
	errControllerConfig              = "cannot get referenced controller config"
	errGetCrd                        = "cannot get crd"
	errDeleteProviderDeployment      = "cannot delete provider package deployment"
	errDeleteProviderSA              = "cannot delete provider package service account"
	errDeleteProviderService         = "cannot delete provider package service"
	errDeleteProviderCertificate     = "cannot delete provider package certificate"
	errDeleteProviderMutateWebhook   = "cannot delete provider package mutate webhook"
	errDeleteProviderValidateWebhook = "cannot delete provider package validate webhook"
	errApplyProviderDeployment       = "cannot apply provider package deployment"
	errApplyProviderCertificate      = "cannot apply provider package certificate"
	errApplyProviderSA               = "cannot apply provider package service account"
	errApplyProviderService          = "cannot apply provider package service"
	errApplyProviderMutateWebhook    = "cannot apply provider package mutate webhook"
	errApplyProviderValidateWebhook  = "cannot apply provider package validate webhook"

	errUnavailableProviderDeployment = "provider package deployment is unavailable"

	errNotIntent              = "not a intent package"
	errNotIntentRevision      = "not a intent revision"
	errDeleteIntentDeployment = "cannot delete intent package deployment"
	errDeleteIntentSA         = "cannot delete intent package service account"
	errDeleteIntentService    = "cannot delete intent package service"

	errApplyIntentDeployment       = "cannot apply intent package deployment"
	errApplyIntentSA               = "cannot apply intent package service account"
	errApplyIntentService          = "cannot apply intent package service"
	errUnavailableIntentDeployment = "intent package deployment is unavailable"
)

// A Hooks performs operations before and after a revision establishes objects.
type Hooks interface {
	// Pre performs operations meant to happen before establishing objects.
	Pre(context.Context, runtime.Object, v1.PackageRevision) error

	// Post performs operations meant to happen after establishing objects.
	Post(context.Context, runtime.Object, v1.PackageRevision, []string) error
}

// IntentHooks performs operations for a Intent package that requires a
// controller before and after the revision establishes objects.
type IntentHooks struct {
	client    resource.ClientApplicator
	namespace string
	log       logging.Logger
}

// NewIntentHooks creates a new IntentHooks.
func NewIntentHooks(client resource.ClientApplicator, namespace string) *IntentHooks {
	return &IntentHooks{
		client:    client,
		namespace: namespace,
	}
}

// Pre cleans up a packaged controller and service account if the revision is
// inactive.
func (h *IntentHooks) Pre(ctx context.Context, pkg runtime.Object, pr v1.PackageRevision) error {
	po, _ := nddpkg.TryConvert(pkg, &pkgmetav1.Intent{})
	pkgIntent, ok := po.(*pkgmetav1.Intent)
	if !ok {
		return errors.New(errNotIntent)
	}

	// TBD updates
	_, ok = pr.(*v1.IntentRevision)
	if !ok {
		return errors.New(errNotIntentRevision)
	}

	//provRev.Status.PermissionRequests = pkgProvider.Spec.Controller.PermissionRequests

	// Do not clean up SA and controller if revision is not inactive.
	if pr.GetDesiredState() != v1.PackageRevisionInactive {
		return nil
	}
	cc, err := h.getControllerConfig(ctx, pr)
	if err != nil {
		return errors.Wrap(err, errControllerConfig)
	}
	svc := buildIntentService(pkgIntent, pr, h.namespace)
	if err := h.client.Delete(ctx, svc); resource.IgnoreNotFound(err) != nil {
		return errors.Wrap(err, errDeleteIntentService)
	}
	s, d := buildIntentDeployment(pkgIntent, pr, cc, h.namespace)
	if err := h.client.Delete(ctx, d); resource.IgnoreNotFound(err) != nil {
		return errors.Wrap(err, errDeleteIntentDeployment)
	}
	if err := h.client.Delete(ctx, s); resource.IgnoreNotFound(err) != nil {
		return errors.Wrap(err, errDeleteIntentSA)
	}
	return nil
}

// Post creates a packaged provider controller and service account if the
// revision is active.
func (h *IntentHooks) Post(ctx context.Context, pkg runtime.Object, pr v1.PackageRevision, crdNames []string) error {
	po, _ := nddpkg.TryConvert(pkg, &pkgmetav1.Intent{})
	pkgIntent, ok := po.(*pkgmetav1.Intent)
	if !ok {
		return errors.New("not a intent package")
	}
	if pr.GetDesiredState() != v1.PackageRevisionActive {
		return nil
	}
	cc, err := h.getControllerConfig(ctx, pr)
	if err != nil {
		return errors.Wrap(err, errControllerConfig)
	}
	svc := buildIntentService(pkgIntent, pr, h.namespace)
	if err := h.client.Apply(ctx, svc); err != nil {
		return errors.Wrap(err, errDeleteIntentService)
	}
	svcMetricHttps := buildIntentMetricServiceHTTPS(pkgIntent, pr, h.namespace)
	if err := h.client.Apply(ctx, svcMetricHttps); err != nil {
		return errors.Wrap(err, errDeleteIntentService)
	}
	svcMetricHttp := buildIntentMetricServiceHTTP(pkgIntent, pr, h.namespace)
	if err := h.client.Apply(ctx, svcMetricHttp); err != nil {
		return errors.Wrap(err, errDeleteIntentService)
	}
	s, d := buildIntentDeployment(pkgIntent, pr, cc, h.namespace)
	if err := h.client.Apply(ctx, s); err != nil {
		return errors.Wrap(err, errApplyProviderSA)
	}
	if err := h.client.Apply(ctx, d); err != nil {
		return errors.Wrap(err, errApplyProviderDeployment)
	}
	pr.SetControllerReference(nddv1.Reference{Name: d.GetName()})

	for _, c := range d.Status.Conditions {
		if c.Type == appsv1.DeploymentAvailable {
			if c.Status == corev1.ConditionTrue {
				return nil
			}
			return errors.Errorf("%s: %s", errUnavailableProviderDeployment, c.Message)
		}
	}
	return nil
}

func (h *IntentHooks) getControllerConfig(ctx context.Context, pr v1.PackageRevision) (*v1.ControllerConfig, error) {
	var cc *v1.ControllerConfig
	if pr.GetControllerConfigRef() != nil {
		cc = &v1.ControllerConfig{}
		if err := h.client.Get(ctx, types.NamespacedName{Name: pr.GetControllerConfigRef().Name}, cc); err != nil {
			return nil, errors.Wrap(err, errControllerConfig)
		}
	}
	return cc, nil
}

// ProviderHooks performs operations for a Provider package that requires a
// controller before and after the revision establishes objects.
type ProviderHooks struct {
	client    resource.ClientApplicator
	namespace string
	log       logging.Logger
}

// NewProviderHooks creates a new ProviderHooks.
func NewProviderHooks(client resource.ClientApplicator, namespace string, l logging.Logger) *ProviderHooks {
	return &ProviderHooks{
		client:    client,
		namespace: namespace,
		log:       l,
	}
}

// Pre cleans up a packaged controller and service account if the revision is
// inactive.
func (h *ProviderHooks) Pre(ctx context.Context, pkg runtime.Object, pr v1.PackageRevision) error {
	po, _ := nddpkg.TryConvert(pkg, &pkgmetav1.Provider{})
	pkgProvider, ok := po.(*pkgmetav1.Provider)
	if !ok {
		return errors.New(errNotProvider)
	}

	// TBD updates
	provRev, ok := pr.(*v1.ProviderRevision)
	if !ok {
		return errors.New(errNotProviderRevision)
	}

	provRev.Status.PermissionRequests = pkgProvider.Spec.Controller.PermissionRequests
	//provRev.Status.Apis = pkgProvider.Spec.Controller.Apis
	//provRev.Status.Pods = pkgProvider.Spec.Controller.Pods

	// Do not clean up SA and controller if revision is not inactive.
	if pr.GetDesiredState() != v1.PackageRevisionInactive {
		return nil
	}
	cc, err := h.getControllerConfig(ctx, pr)
	if err != nil {
		return errors.Wrap(err, errControllerConfig)
	}
	// We do not have o delete the package since it has a common name; it will be deleted because the owner reference deals with that
	/*
		h.log.Debug("pkgProvide deleter", "pkgProvider", pkgProvider, "desired statis", pr.GetDesiredState())
		// set the namepsace to the one of ndd-core -> ndd-system
		pkgProvider.SetNamespace(h.namespace)
		if err := h.client.Delete(ctx, pkgProvider); resource.IgnoreNotFound(err) != nil {
			return errors.Wrap(err, "error delete metav1 provider")
		}
	*/
	/*
		svc := buildProviderService(pkgProvider, pr, h.namespace)
		if err := h.client.Delete(ctx, svc); resource.IgnoreNotFound(err) != nil {
			return errors.Wrap(err, errDeleteProviderService)
		}
	*/
	//svcProfile := buildProviderProfileService(pkgProvider, pr, h.namespace)
	//if err := h.client.Delete(ctx, svcProfile); err != nil {
	//	return errors.Wrap(err, errDeleteProviderService)
	//}
	/*
		svcWebHook := buildProviderWebhookService(pkgProvider, pr, h.namespace)
		if err := h.client.Delete(ctx, svcWebHook); err != nil {
			return errors.Wrap(err, errDeleteProviderService)
		}
		certWebHook := buildProviderWebhookCertificate(pkgProvider, pr, h.namespace)
		if err := h.client.Delete(ctx, certWebHook); err != nil {
			return errors.Wrap(err, errDeleteProviderCertificate)
		}
		mutateWebHook := buildProviderWebhookMutate(pkgProvider, pr, h.namespace)
		if err := h.client.Delete(ctx, mutateWebHook); err != nil {
			return errors.Wrap(err, errDeleteProviderMutateWebhook)
		}
		validateWebHook := buildProviderWebhookValidate(pkgProvider, pr, h.namespace)
		if err := h.client.Delete(ctx, validateWebHook); err != nil {
			return errors.Wrap(err, errDeleteProviderValidateWebhook)
		}
	*/
	s, d := buildProviderDeployment(pkgProvider, pr, cc, h.namespace)
	if err := h.client.Delete(ctx, d); resource.IgnoreNotFound(err) != nil {
		return errors.Wrap(err, errDeleteProviderDeployment)
	}
	if err := h.client.Delete(ctx, s); resource.IgnoreNotFound(err) != nil {
		return errors.Wrap(err, errDeleteProviderSA)
	}
	return nil
}

// Post creates a packaged provider controller and service account if the
// revision is active.
func (h *ProviderHooks) Post(ctx context.Context, pkg runtime.Object, pr v1.PackageRevision, crdNames []string) error {
	po, _ := nddpkg.TryConvert(pkg, &pkgmetav1.Provider{})
	pkgProvider, ok := po.(*pkgmetav1.Provider)
	if !ok {
		return errors.New("not a provider package")
	}

	// return if the desired status is not active
	if pr.GetDesiredState() != v1.PackageRevisionActive {
		return nil
	}
	cc, err := h.getControllerConfig(ctx, pr)
	if err != nil {
		return errors.Wrap(err, errControllerConfig)
	}

	// only create when not found, since we want to allow human or automated pipelines to update the CR
	// we just create it
	pkgMeta := &pkgmetav1.Provider{}
	if err := h.client.Get(ctx, types.NamespacedName{Namespace: h.namespace, Name: pkgProvider.GetName()}, pkgMeta); err != nil {
		if resource.IgnoreNotFound(err) != nil {
			return errors.Wrap(err, "error get metav1 provider")
		}
		pkgprov := buildProviderPackage(pkgProvider, pr, h.namespace)
		if err := h.client.Apply(ctx, pkgprov); err != nil {
			return errors.Wrap(err, "error create metav1 provider")
		}
	} else {
		// check the owner reference and if it differs we update it
		for _, ref := range pkgMeta.GetOwnerReferences() {
			if ref.UID != pr.GetUID() {
				pkgprov := buildProviderPackage(pkgProvider, pr, h.namespace)
				if err := h.client.Apply(ctx, pkgprov); err != nil {
					return errors.Wrap(err, "error create metav1 provider")
				}
			}
		}
	}
	h.log.Debug("pkgProvider", "pkgMeta", pkgMeta.Spec)

	/*
		crds, err := h.getCrds(ctx, crdNames)
		if err != nil {
			return errors.Wrap(err, errGetCrd)
		}
	*/
	/*
		svcGnmi := buildProviderService(pkgProvider, pr, h.namespace)
		if err := h.client.Apply(ctx, svcGnmi); err != nil {
			return errors.Wrap(err, errDeleteProviderService)
		}
	*/
	/*
		certGnmi := buildProviderGnmiCertificate(pkgProvider, pr, h.namespace)
		if err := h.client.Apply(ctx, certGnmi); err != nil {
			return errors.Wrap(err, errApplyProviderCertificate)
		}
	*/
	svcMetricHttps := buildProviderMetricServiceHTTPS(pkgProvider, pr, h.namespace)
	if err := h.client.Apply(ctx, svcMetricHttps); err != nil {
		return errors.Wrap(err, errApplyProviderService)
	}
	svcProfile := buildProviderProfileService(pkgProvider, pr, h.namespace)
	if err := h.client.Apply(ctx, svcProfile); err != nil {
		return errors.Wrap(err, errApplyProviderService)
	}
	/*
		if len(crds) == 1 {
			svcWebHook := buildProviderWebhookService(pkgProvider, pr, h.namespace)
			if err := h.client.Apply(ctx, svcWebHook); err != nil {
				return errors.Wrap(err, errApplyProviderService)
			}
			certWebHook := buildProviderWebhookCertificate(pkgProvider, pr, h.namespace)
			if err := h.client.Apply(ctx, certWebHook); err != nil {
				return errors.Wrap(err, errApplyProviderCertificate)
			}
			mutateWebHook := buildProviderWebhookMutate(pkgProvider, pr, h.namespace, crds[0])
			if err := h.client.Apply(ctx, mutateWebHook); err != nil {
				return errors.Wrap(err, errApplyProviderMutateWebhook)
			}
			validateWebHook := buildProviderWebhookValidate(pkgProvider, pr, h.namespace, crds[0])
			if err := h.client.Apply(ctx, validateWebHook); err != nil {
				return errors.Wrap(err, errApplyProviderValidateWebhook)
			}
		}
	*/
	s, d := buildProviderDeployment(pkgProvider, pr, cc, h.namespace)
	if err := h.client.Apply(ctx, s); err != nil {
		return errors.Wrap(err, errApplyProviderSA)
	}
	if err := h.client.Apply(ctx, d); err != nil {
		return errors.Wrap(err, errApplyProviderDeployment)
	}
	pr.SetControllerReference(nddv1.Reference{Name: d.GetName()})

	for _, c := range d.Status.Conditions {
		if c.Type == appsv1.DeploymentAvailable {
			if c.Status == corev1.ConditionTrue {
				return nil
			}
			return errors.Errorf("%s: %s", errUnavailableProviderDeployment, c.Message)
		}
	}
	return nil
}

func (h *ProviderHooks) getControllerConfig(ctx context.Context, pr v1.PackageRevision) (*v1.ControllerConfig, error) {
	var cc *v1.ControllerConfig
	if pr.GetControllerConfigRef() != nil {
		cc = &v1.ControllerConfig{}
		if err := h.client.Get(ctx, types.NamespacedName{Name: pr.GetControllerConfigRef().Name}, cc); err != nil {
			return nil, errors.Wrap(err, errControllerConfig)
		}
	}
	return cc, nil
}

func (h *ProviderHooks) getCrds(ctx context.Context, crdNames []string) ([]*extv1.CustomResourceDefinition, error) {
	crds := []*extv1.CustomResourceDefinition{}
	for _, crdName := range crdNames {
		crd := &extv1.CustomResourceDefinition{}
		if err := h.client.Get(ctx, types.NamespacedName{Name: crdName}, crd); err != nil {
			return nil, errors.Wrap(err, errGetCrd)
		}
		crds = append(crds, crd)
	}
	return crds, nil
}

// NopHooks performs no operations.
type NopHooks struct{}

// NewNopHooks creates a hook that does nothing.
func NewNopHooks() *NopHooks {
	return &NopHooks{}
}

// Pre does nothing and returns nil.
func (h *NopHooks) Pre(context.Context, runtime.Object, v1.PackageRevision) error {
	return nil
}

// Post does nothing and returns nil.
func (h *NopHooks) Post(context.Context, runtime.Object, v1.PackageRevision, []string) error {
	return nil
}
