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
	nddv1 "github.com/yndd/ndd-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProviderSpec specifies details about a request to install a provider
type ProviderSpec struct {
	PackageSpec `json:",inline"`

	// ControllerRef references a ControllerConfig resource that will be
	// used to configure the packaged controller Deployment.
	// +optional
	ControllerReference *nddv1.Reference `json:"controllerRef,omitempty"`
}

// ProviderStatus defines the observed state of Provider
type ProviderStatus struct {
	nddv1.ConditionedStatus `json:",inline"`
	PackageStatus           `json:",inline"`
}

// +kubebuilder:object:root=true
// +genclient
// +genclient:nonNamespaced

// Provider is the CRD type for a request to add a provider to Network Device Driver..
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="INSTALLED",type="string",JSONPath=".status.conditions[?(@.kind=='PackageInstalled')].status"
// +kubebuilder:printcolumn:name="HEALTHY",type="string",JSONPath=".status.conditions[?(@.kind=='PackageHealthy')].status"
// +kubebuilder:printcolumn:name="PACKAGE",type="string",JSONPath=".spec.package"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={ndd,pkg},shortName=pvd
type Provider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderSpec   `json:"spec,omitempty"`
	Status ProviderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type ProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Provider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Provider{}, &ProviderList{})
}

// +kubebuilder:object:root=true
// +genclient
// +genclient:nonNamespaced

// A ProviderRevision that has been added to Network device Driver.
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="HEALTHY",type="string",JSONPath=".status.conditions[?(@.kind=='PackageHealthy')].status"
// +kubebuilder:printcolumn:name="REVISION",type="string",JSONPath=".spec.revision"
// +kubebuilder:printcolumn:name="PKGIMAGE",type="string",JSONPath=".spec.packageImage"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".spec.desiredState"
// +kubebuilder:printcolumn:name="DEP-FOUND",type="string",JSONPath=".status.foundDependencies"
// +kubebuilder:printcolumn:name="DEP-INSTALLED",type="string",JSONPath=".status.installedDependencies"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={ndd,pkg},shortName=prov
type ProviderRevision struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PackageRevisionSpec   `json:"spec,omitempty"`
	Status PackageRevisionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProviderRevisionList contains a list of ProviderRevision.
type ProviderRevisionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderRevision `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProviderRevision{}, &ProviderRevisionList{})
}
