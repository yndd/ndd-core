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

	ndddvrv1 "github.com/yndd/ndd-core/apis/dvr/v1"
	nddv1 "github.com/yndd/ndd-runtime/apis/common/v1"

	"github.com/pkg/errors"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-runtime/pkg/resource"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	errNotProvider              = "not a provider package"
	errNotProviderRevision      = "not a provider revision"
	errControllerConfig         = "cannot get referenced controller config"
	errDeleteDeployment         = "cannot delete device driver deployment"
	errDeleteServiceAccount     = "cannot delete device driver service account"
	errDeleteConfigMap          = "cannot delete device driver config map"
	errDeleteService            = "cannot delete device driver service"
	errDeleteClusterRoleBinding = "cannot delete device driver cluster role binding"
	errApplyDeployment          = "cannot apply device driver deployment"
	errApplyServiceAccount      = "cannot apply device driver service account"
	errApplyConfigMap           = "cannot apply device driver config map"
	errApplyService             = "cannot apply device driver service"
	errAppyClusterRoleBinding   = "cannot apply device driver cluster role binding"
	errUnavailableDeployment    = "device driver deployment is unavailable"
)

// A Hooks performs operations to deploy the device driver for the network node.
type Hooks interface {
	// Deploy performs operations to deploy the device driver for the network node
	Deploy(ctx context.Context, nn ndddvrv1.Nn, c *corev1.Container) error

	// Destroy performs operations to destroy the device driver for the network node
	Destroy(ctx context.Context, nn ndddvrv1.Nn, c *corev1.Container) error
}

// DeviceDriverHooks performs operations to deploy the device driver.
type DeviceDriverHooks struct {
	client    resource.ClientApplicator
	log       logging.Logger
	namespace string
}

// NewPDeviceDriverHooks creates a new DeviceDriverHooks.
func NewDeviceDriverHooks(client resource.ClientApplicator, log logging.Logger, namespace string) *DeviceDriverHooks {
	return &DeviceDriverHooks{
		client:    client,
		log:       log,
		namespace: namespace,
	}
}

// Deploy performs operations to deploy the device driver for the network node
func (h *DeviceDriverHooks) Deploy(ctx context.Context, nn ndddvrv1.Nn, c *corev1.Container) error {
	cm := buildConfigMap(nn, h.namespace)
	if err := h.client.Apply(ctx, cm); err != nil {
		return errors.Wrap(err, errApplyConfigMap)
	}

	s := buildService(nn, h.namespace)
	if err := h.client.Apply(ctx, s); err != nil {
		return errors.Wrap(err, errApplyService)
	}

	sa := buildServiceAccount(nn, h.namespace)
	if err := h.client.Apply(ctx, sa); err != nil {
		return errors.Wrap(err, errApplyServiceAccount)
	}

	d := buildDeployment(nn, c, h.namespace)
	if err := h.client.Apply(ctx, d); err != nil {
		return errors.Wrap(err, errApplyDeployment)
	}

	b := buildClusterRoleBinding(nn, h.namespace)
	if err := h.client.Apply(ctx, b); err != nil {
		return errors.Wrap(err, errAppyClusterRoleBinding)
	}
	nn.SetControllerReference(nddv1.Reference{Name: d.GetName()})

	for _, c := range d.Status.Conditions {
		if c.Type == appsv1.DeploymentAvailable {
			if c.Status == corev1.ConditionTrue {
				return nil
			}
			return errors.Errorf("%s: %s", errUnavailableDeployment, c.Message)
		}
	}
	return nil
}

// Destroy performs operations to destroy the device driver for the network node
func (h *DeviceDriverHooks) Destroy(ctx context.Context, nn ndddvrv1.Nn, c *corev1.Container) error {
	cm := buildConfigMap(nn, h.namespace)
	if err := h.client.Delete(ctx, cm); err != nil {
		return errors.Wrap(err, errDeleteConfigMap)
	}

	s := buildService(nn, h.namespace)
	if err := h.client.Delete(ctx, s); err != nil {
		return errors.Wrap(err, errDeleteService)
	}

	sa := buildServiceAccount(nn, h.namespace)
	if err := h.client.Delete(ctx, sa); err != nil {
		return errors.Wrap(err, errDeleteServiceAccount)
	}

	d := buildDeployment(nn, c, h.namespace)
	if err := h.client.Delete(ctx, d); err != nil {
		return errors.Wrap(err, errDeleteDeployment)
	}

	b := buildClusterRoleBinding(nn, h.namespace)
	if err := h.client.Delete(ctx, b); err != nil {
		return errors.Wrap(err, errDeleteClusterRoleBinding)
	}
	nn.SetControllerReference(nddv1.Reference{Name: d.GetName()})

	for _, c := range d.Status.Conditions {
		if c.Type == appsv1.DeploymentAvailable {
			if c.Status == corev1.ConditionTrue {
				return nil
			}
			return errors.Errorf("%s: %s", errUnavailableDeployment, c.Message)
		}
	}
	return nil
}

// NopHooks performs no operations.
type NopHooks struct{}

// NewNopHooks creates a hook that does nothing.
func NewNopHooks() *NopHooks {
	return &NopHooks{}
}

// Deploy does nothing and returns nil.
func (h *NopHooks) Deploy(ctx context.Context, nn ndddvrv1.Nn, c *corev1.Container) error {
	return nil
}

// Destroy does nothing and returns nil.
func (h *NopHooks) Destroy(ctx context.Context, nn ndddvrv1.Nn, c *corev1.Container) error {
	return nil
}
