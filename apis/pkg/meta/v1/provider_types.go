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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ControllerType string

const (
	ControllerTypeController ControllerType = "controller"
	ControllerTypeIntent     ControllerType = "intent"
)

// ProviderSpec specifies the configuration of a Provider.
type ProviderSpec struct {
	// Type is the type of provider
	// +kubebuilder:default=controller
	Type ControllerType `json:"type,omitempty"`

	// Configuration for the packaged Provider's controller.
	Controller ControllerSpec `json:"controller"`

	MetaSpec `json:",inline"`
}

/*
type Api struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}
*/

// ControllerSpec specifies the configuration for the packaged Provider
// controller.
type ControllerSpec struct {
	// Image is the packaged Provider controller image.
	Image string `json:"image"`

	// PermissionRequests for RBAC rules required for this provider's controller
	// to function. The RBAC manager is responsible for assessing the requested
	// permissions.
	// +optional
	PermissionRequests []rbacv1.PolicyRule `json:"permissionRequests,omitempty"`

	Pods []PodSpec `json:"pods,omitempty"`

	//Apis []Api `json:"apis,omitempty"`
}

type DeploymentType string

const (
	DeploymentTypeStatefulset DeploymentType = "statefulset"
	DeploymentTypeDeployment  DeploymentType = "deployment"
)

type PodSpec struct {
	// Name of the pod
	Name string `json:"name"`

	// Type is the type of the deployment
	// +kubebuilder:default=statefulset
	Type DeploymentType `json:"type,omitempty"`

	// PermissionRequests for RBAC rules required for this provider's controller
	// to function. The RBAC manager is responsible for assessing the requested
	// permissions.
	// +optional
	PermissionRequests []rbacv1.PolicyRule `json:"permissionRequests,omitempty"`

	// Containers identifies the containers in the pod
	Containers []ContainerSpec `json:"containers,omitempty"`
}

type ContainerSpec struct {
	// Provide the container info
	Container *corev1.Container `json:"container,omitempty"`

	// Extras is certificates, volumes, webhook, etc
	Extras []Extras `json:"extras,omitempty"`
}

type Extras struct {
	Name        string `json:"name"`
	Webhook     bool   `json:"webhook,omitempty"`
	Certificate bool   `json:"certificate,omitempty"`
	Service     bool   `json:"service,omitempty"`
	Volume      bool   `json:"volume,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

// A Provider is the description of a Ndd Provider package.
type Provider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ProviderSpec `json:"spec"`
}

//+kubebuilder:object:root=true

// A ProviderList is the description of a Ndd Provider package.
type ProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Provider `json:"items"`
}

// Hub marks this type as the conversion hub.
func (p *Provider) Hub() {}

func init() {
	SchemeBuilder.Register(&Provider{}, &ProviderList{})
}
