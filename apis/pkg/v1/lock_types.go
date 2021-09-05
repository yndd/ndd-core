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
	pkgmetav1 "github.com/yndd/ndd-core/apis/pkg/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LockPackage is a package that is in the lock.
type LockPackage struct {
	// Name corresponds to the name of the package revision for this package.
	Name string `json:"name"`

	// Type is the type of package. Can be Provider, maybe later other types can be created
	Type pkgmetav1.PackageType `json:"type"`

	// Source is the OCI image name without a tag or digest.
	Source string `json:"source"`

	// Version is the tag or digest of the OCI image.
	Version string `json:"version"`

	// Dependencies are the list of dependencies of this package. The order of
	// the dependencies will dictate the order in which they are resolved.
	Dependencies []pkgmetav1.Dependency `json:"dependencies"`
}

// +kubebuilder:object:root=true
// +genclient
// +genclient:nonNamespaced

// LockSpec is the CRD type that tracks package dependencies.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={ndd,pkg},shortName=lock
type Lock struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Packages []LockPackage `json:"packages,omitempty"`
}

// +kubebuilder:object:root=true

// LockList contains a list of Lock.
type LockList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Lock `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Lock{}, &LockList{})
}
