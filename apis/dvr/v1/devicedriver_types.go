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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeviceDriverSpec defines the desired state of DeviceDriver
type DeviceDriverSpec struct {
	// Container defines the container parameters for the device driver
	Container *corev1.Container `json:"container,omitempty"`
}

//+kubebuilder:object:root=true

// DeviceDriver is the Schema for the devicedrivers API
type DeviceDriver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DeviceDriverSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// DeviceDriverList contains a list of DeviceDriver
type DeviceDriverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeviceDriver `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DeviceDriver{}, &DeviceDriverList{})
}
