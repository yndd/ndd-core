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
	"reflect"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ControllerType string

const (
	ControllerTypeController ControllerType = "controller"
	ControllerTypeIntent     ControllerType = "intent"
	ControllerTypeProvider   ControllerType = "provider"
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

	//Pods []PodSpec `json:"pods,omitempty"`
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

// Provider type metadata.
var (
	ProviderKind             = reflect.TypeOf(Provider{}).Name()
	ProviderGroupKind        = schema.GroupKind{Group: Group, Kind: ProviderKind}.String()
	ProviderKindAPIVersion   = ProviderKind + "." + GroupVersion.String()
	ProviderGroupVersionKind = GroupVersion.WithKind(ProviderKind)
)
