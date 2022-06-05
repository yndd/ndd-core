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
	"fmt"
	"reflect"
	"strings"

	nddv1 "github.com/yndd/ndd-runtime/apis/common/v1"
	targetv1 "github.com/yndd/target/apis/target/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func GetServiceName(prefix, name string) string {
	return strings.Join([]string{prefix, name}, "-")
}

func GetServiceTag(namespace, name string) []string {
	return []string{fmt.Sprintf("%s=%s/%s", serviceTag, namespace, name)}
}

func GetTargetTag(namespace, name string) []string {
	return []string{fmt.Sprintf("%s=%s", targetService, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}.String())}
}

type ServiceInfo struct {
	ServiceName string
	Kind        Kind
}

func (pc *CompositeProvider) GetPodServiceInfo(podName string, k Kind) *ServiceInfo {
	return &ServiceInfo{
		ServiceName: GetServiceName(pc.Name, podName),
		Kind:        k,
	}
}

func (pc *CompositeProvider) GetTargetServiceInfo() *ServiceInfo {
	return &ServiceInfo{
		ServiceName: GetServiceName(pc.Name, targetService),
		Kind:        KindNone,
	}
}

func (pc *CompositeProvider) GetServicesInfo() []*ServiceInfo {
	services := make([]*ServiceInfo, 0, len(pc.Spec.Packages)+1)
	for _, pkg := range pc.Spec.Packages {
		services = append(services, &ServiceInfo{
			ServiceName: GetServiceName(pc.Name, pkg.Name),
			Kind:        pkg.Kind,
		})
	}
	return services
}

func (pc *CompositeProvider) GetAllServicesInfo() []*ServiceInfo {
	services := pc.GetServicesInfo()
	services = append(services, pc.GetTargetServiceInfo())
	return services
}

func (pc *CompositeProvider) GetServicesInfoByKind(kind Kind) []*ServiceInfo {
	services := make([]*ServiceInfo, 0, len(pc.Spec.Packages)+1)
	for _, pkg := range pc.Spec.Packages {
		if pkg.Kind == kind {
			services = append(services, &ServiceInfo{
				ServiceName: GetServiceName(pc.Name, pkg.Name),
				Kind:        pkg.Kind,
			})
			// break not added to make it more generic in the future if multiple pods have the same kind
		}
	}
	if kind == KindWorker {
		services = append(services, pc.GetTargetServiceInfo())
	}
	return services
}

type Kind string

const (
	KindNone       Kind = ""
	KindWorker     Kind = "worker"
	KindReconciler Kind = "reconciler"
)

// ControllerSpec specifies the configuration of a Controller.
type CompositeProviderSpec struct {
	// VendorType specifies the vendor of the provider composite
	//+kubebuilder:validation:Enum=unknown;nokiaSRL;nokiaSROS;
	VendorType targetv1.VendorType `json:"vendorType,omitempty"`
	// Packages define the package specification used for creating the provider
	Packages []PackageSpec `json:"packages,omitempty"`
}

// CompositeProviderStatus defines the observed state of CompositeProvider
type CompositeProviderStatus struct {
	nddv1.ConditionedStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +genclient
// +genclient:nonNamespaced

// A CompositeProvider provides the definition of a CompositeProvider configuration.
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="INSTALLED",type="string",JSONPath=".status.conditions[?(@.kind=='PackageInstalled')].status"
// +kubebuilder:printcolumn:name="HEALTHY",type="string",JSONPath=".status.conditions[?(@.kind=='PackageHealthy')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={ndd,pkg}
type CompositeProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CompositeProviderSpec   `json:"spec"`
	Status CompositeProviderStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// A CompositeProviderList provides the list of CompositeProvider.
type CompositeProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []CompositeProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CompositeProvider{}, &CompositeProviderList{})
}

// CompositeProvider type metadata.
var (
	CompositeProviderKind             = reflect.TypeOf(CompositeProvider{}).Name()
	CompositeProviderGroupKind        = schema.GroupKind{Group: Group, Kind: CompositeProviderKind}.String()
	CompositeProviderKindAPIVersion   = ProviderKind + "." + GroupVersion.String()
	CompositeProviderGroupVersionKind = GroupVersion.WithKind(CompositeProviderKind)
)
