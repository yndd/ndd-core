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
	"github.com/yndd/ndd-runtime/pkg/resource"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const (
	errNotProvider                   = "not a provider package"
	errNotProviderRevision           = "not a provider revision"
	errControllerConfig              = "cannot get referenced controller config"
	errDeleteProviderDeployment      = "cannot delete provider package deployment"
	errDeleteProviderSA              = "cannot delete provider package service account"
	errApplyProviderDeployment       = "cannot apply provider package deployment"
	errApplyProviderSA               = "cannot apply provider package service account"
	errUnavailableProviderDeployment = "provider package deployment is unavailable"
)

// A Hooks performs operations before and after a revision establishes objects.
type Hooks interface {
	// Pre performs operations meant to happen before establishing objects.
	Pre(context.Context, runtime.Object, v1.PackageRevision) error

	// Post performs operations meant to happen after establishing objects.
	Post(context.Context, runtime.Object, v1.PackageRevision) error
}

// ProviderHooks performs operations for a provider package that requires a
// controller before and after the revision establishes objects.
type ProviderHooks struct {
	client    resource.ClientApplicator
	namespace string
}

// NewProviderHooks creates a new ProviderHooks.
func NewProviderHooks(client resource.ClientApplicator, namespace string) *ProviderHooks {
	return &ProviderHooks{
		client:    client,
		namespace: namespace,
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
	_, ok = pr.(*v1.ProviderRevision)
	if !ok {
		return errors.New(errNotProviderRevision)
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
func (h *ProviderHooks) Post(ctx context.Context, pkg runtime.Object, pr v1.PackageRevision) error {
	po, _ := nddpkg.TryConvert(pkg, &pkgmetav1.Provider{})
	pkgProvider, ok := po.(*pkgmetav1.Provider)
	if !ok {
		return errors.New("not a provider package")
	}
	if pr.GetDesiredState() != v1.PackageRevisionActive {
		return nil
	}
	cc, err := h.getControllerConfig(ctx, pr)
	if err != nil {
		return errors.Wrap(err, errControllerConfig)
	}
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
func (h *NopHooks) Post(context.Context, runtime.Object, v1.PackageRevision) error {
	return nil
}
