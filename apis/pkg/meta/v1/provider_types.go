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

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ServiceDiscoveryType string

const (
	ServiceDiscoveryTypeConsul ServiceDiscoveryType = "consul"
	ServiceDiscoveryTypeK8s    ServiceDiscoveryType = "k8s"
)

type DeploymentType string

const (
	DeploymentTypeStatefulset DeploymentType = "statefulset"
	DeploymentTypeDeployment  DeploymentType = "deployment"
)

// ProviderSpec specifies the configuration of a Provider.
type ProviderSpec struct {
	// ServiceDiscovery is the type of service discovery
	// +kubebuilder:validation:Enum=`consul`;`k8s`
	// +kubebuilder:default=consul
	ServiceDiscovery ServiceDiscoveryType `json:"service-discovery,omitempty"`
	// ServiceDiscoverylNamespace is the name of the service discovery namespace
	// +kubebuilder:default=consul
	ServiceDiscoveryNamespace string `json:"service-discovery-namespace,omitempty"`
	// pods define the pod specification used by the controller for LCM/resource allocation
	Pod *PodSpec `json:"pods,omitempty"`

	MetaSpec `json:",inline"`
}

type PodSpec struct {
	// Name of the pod
	Name string `json:"name,omitempty"`

	// Type is the type of the deployment
	// +kubebuilder:validation:Enum=`statefulset`;`deployment`
	// +kubebuilder:default=deployment
	Type DeploymentType `json:"type,omitempty"`

	// Replicas defines the amount of replicas expected
	// +kubebuilder:default=1
	Replicas int `json:"replicas,omitempty"`

	// MaxReplicas defines the max expected replications of this pod
	// +kubebuilder:default=8
	MaxReplicas int `json:"max-replicas,omitempty"`

	// MaxJobNumber indication on how many jobs a given pods should hold
	MaxJobNumber int `json:"max-job-number,omitempty"`

	// PermissionRequests for RBAC rules required for this controller
	// to function. The RBAC manager is responsible for assessing the requested
	// permissions.
	// +optional
	PermissionRequests []rbacv1.PolicyRule `json:"permission-requests,omitempty"`

	// Containers identifies the containers in the pod
	Containers []*ContainerSpec `json:"containers,omitempty"`
}

type ContainerSpec struct {
	// Provide the container info
	Container *corev1.Container `json:"container,omitempty"`

	// Extras is certificates, volumes, webhook, etc
	Extras []*Extras `json:"extras,omitempty"`
}

type Extras struct {
	Name        string `json:"name"`
	Webhook     bool   `json:"webhook,omitempty"`
	Certificate bool   `json:"certificate,omitempty"`
	Service     bool   `json:"service,omitempty"`
	Volume      bool   `json:"volume,omitempty"`
	Port        uint32 `json:"port,omitempty"`
	TargetPort  uint32 `json:"target-port,omitempty"`
	Protocol    string `json:"protocol,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

// A Provider is the description of a Ndd Provider package.
type Provider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ProviderSpec `json:"spec"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

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
