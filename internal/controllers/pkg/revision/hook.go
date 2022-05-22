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
	pkgv1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/yndd/ndd-core/internal/nddpkg"
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
	errApplyProviderStatefulset      = "cannot apply provider package statefulset"
	errApplyProviderCertificate      = "cannot apply provider package certificate"
	errApplyProviderServiceAccount   = "cannot apply provider package service account"
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
	Pre(context.Context, runtime.Object, pkgv1.PackageRevision, []string) error

	// Post performs operations meant to happen after establishing objects.
	Post(context.Context, runtime.Object, pkgv1.PackageRevision, []string) error
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
func (h *ProviderHooks) Pre(ctx context.Context, pkg runtime.Object, pr pkgv1.PackageRevision, crdNames []string) error {
	log := h.log.WithValues("package", pkg.GetObjectKind(), "pr", pr.GetName())
	po, _ := nddpkg.TryConvert(pkg, &pkgmetav1.Provider{})
	pmp, ok := po.(*pkgmetav1.Provider)
	if !ok {
		return errors.New(errNotProvider)
	}

	// TBD updates
	provRev, ok := pr.(*pkgv1.ProviderRevision)
	if !ok {
		return errors.New(errNotProviderRevision)
	}

	log.Debug("permission requests", "meta spec", pmp.Spec, "revision status", provRev.Status)
	provRev.Status.PermissionRequests = pmp.Spec.Pod.PermissionRequests

	// Do not clean up if revision is active.
	if pr.GetDesiredState() == pkgv1.PackageRevisionActive {
		return nil
	}
	log.Debug("desired state", "state", pr.GetDesiredState())
	/*
		cc, err := h.getController(ctx, pr)
		if err != nil {
			return errors.Wrap(err, errControllerConfig)
		}
	*/
	// We do not have to delete the package since it has a common name;
	// it will be deleted because the owner reference deals with that
	switch pmp.Spec.Pod.Type {
	case pkgmetav1.DeploymentTypeDeployment:
		d := renderProviderDeployment(pmp, pmp.Spec.Pod, pr, &Options{})
		if err := h.client.Delete(ctx, d); resource.IgnoreNotFound(err) != nil {
			return errors.Wrap(err, errDeleteProviderDeployment)
		}
		sa := renderServiceAccount(pmp, pmp.Spec.Pod, pr)
		if err := h.client.Delete(ctx, sa); resource.IgnoreNotFound(err) != nil {
			return errors.Wrap(err, errDeleteProviderSA)
		}
	case pkgmetav1.DeploymentTypeStatefulset:
		d := renderProviderStatefulSet(pmp, pmp.Spec.Pod, pr, &Options{})
		if err := h.client.Delete(ctx, d); resource.IgnoreNotFound(err) != nil {
			return errors.Wrap(err, errDeleteProviderDeployment)
		}
		sa := renderServiceAccount(pmp, pmp.Spec.Pod, pr)
		if err := h.client.Delete(ctx, sa); resource.IgnoreNotFound(err) != nil {
			return errors.Wrap(err, errDeleteProviderSA)
		}
	}

	return nil
}

// Post creates a packaged provider controller and service account if the
// revision is active.
func (h *ProviderHooks) Post(ctx context.Context, pkg runtime.Object, pr pkgv1.PackageRevision, crdNames []string) error {
	log := h.log.WithValues("package", pkg.GetObjectKind(), "pr", pr.GetName())
	po, _ := nddpkg.TryConvert(pkg, &pkgmetav1.Provider{})
	pmp, ok := po.(*pkgmetav1.Provider)
	if !ok {
		return errors.New("not a provider package")
	}

	// return if the desired status is not active
	if pr.GetDesiredState() != pkgv1.PackageRevisionActive {
		return nil
	}

	crds, err := h.getCrds(ctx, crdNames)
	if err != nil {
		return errors.Wrap(err, errGetCrd)
	}

	var grpcServiceName string
	for _, c := range pmp.Spec.Pod.Containers {
		log.Debug("extras", "container", c.Container.Name, "extras", c.Extras)
		for _, extra := range c.Extras {
			if extra.Certificate {
				// deploy a certificate
				cert := renderCertificate(pmp, pmp.Spec.Pod, c, extra, pr)
				if err := h.client.Apply(ctx, cert); err != nil {
					return errors.Wrap(err, errApplyProviderCertificate)
				}
			}
			if extra.Service {
				// deploy a service
				s := renderService(pmp, pmp.Spec.Pod, c, extra, pr)
				if err := h.client.Apply(ctx, s); err != nil {
					return errors.Wrap(err, errApplyProviderService)
				}
				if extra.Name == "grpc" {
					grpcServiceName = s.Name
				}
			}
			if extra.Webhook {
				// deploy a mutating webhook
				whMutate := renderWebhookMutate(pmp, pmp.Spec.Pod, c, extra, pr, crds)
				if err := h.client.Apply(ctx, whMutate); err != nil {
					return errors.Wrap(err, errApplyProviderMutateWebhook)
				}
				// deploy a validating webhook
				whValidate := renderWebhookValidate(pmp, pmp.Spec.Pod, c, extra, pr, crds)
				if err := h.client.Apply(ctx, whValidate); err != nil {
					return errors.Wrap(err, errApplyProviderValidateWebhook)
				}
			}
		}
	}

	switch pmp.Spec.Pod.Type {
	case pkgmetav1.DeploymentTypeDeployment:
		d := renderProviderDeployment(pmp, pmp.Spec.Pod, pr, &Options{
			grpcServiceName: grpcServiceName,
		})
		if err := h.client.Apply(ctx, d); err != nil {
			return errors.Wrap(err, errApplyProviderDeployment)
		}
		sa := renderServiceAccount(pmp, pmp.Spec.Pod, pr)
		if err := h.client.Apply(ctx, sa); err != nil {
			return errors.Wrap(err, errApplyProviderServiceAccount)
		}
		for _, c := range d.Status.Conditions {
			if c.Type == appsv1.DeploymentAvailable {
				if c.Status == corev1.ConditionTrue {
					return nil
				}
				return errors.Errorf("%s: %s", errUnavailableProviderDeployment, c.Message)
			}
		}
	case pkgmetav1.DeploymentTypeStatefulset:
		cc, err := h.getController(ctx, pr)
		if err != nil {
			return errors.Wrap(err, errControllerConfig)
		}
		s := renderProviderStatefulSet(pmp, pmp.Spec.Pod, pr, &Options{
			serviceDiscoveryInfo: cc.GetServicesInfoByKind(pr.GetRevisionKind()),
			grpcServiceName:      grpcServiceName,
		})
		if err := h.client.Apply(ctx, s); err != nil {
			return errors.Wrap(err, errApplyProviderStatefulset)
		}
		sa := renderServiceAccount(pmp, pmp.Spec.Pod, pr)
		if err := h.client.Apply(ctx, sa); err != nil {
			return errors.Wrap(err, errApplyProviderServiceAccount)
		}
	}

	return nil
}

func (h *ProviderHooks) getController(ctx context.Context, pr pkgv1.PackageRevision) (*pkgv1.CompositeProvider, error) {
	var cc *pkgv1.CompositeProvider
	if pr.GetControllerRef() != nil {
		cc = &pkgv1.CompositeProvider{}
		if err := h.client.Get(ctx, types.NamespacedName{Name: pr.GetControllerRef().Name}, cc); err != nil {
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
func (h *NopHooks) Pre(context.Context, runtime.Object, pkgv1.PackageRevision, []string) error {
	return nil
}

// Post does nothing and returns nil.
func (h *NopHooks) Post(context.Context, runtime.Object, pkgv1.PackageRevision, []string) error {
	return nil
}
