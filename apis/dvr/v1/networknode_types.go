/*
Copyright 2021 Wim Henderickx.

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

// NetworkNodeSpec defines the desired state of NetworkNode
type NetworkNodeSpec struct {
	// Target defines the details how we connect to the network device
	Target *TargetDetails `json:"target"`

	// DeviceDriver defines the device driver details to connect to the network device
	// +optional
	// +kubebuilder:default=gnmi
	DeviceDriverKind *DeviceDriverKind `json:"deviceDriverKind,omitempty"`

	// GrpcServerPort defines the grpc server port to connect to the device driver
	// from the network device provider
	// +optional
	// +kubebuilder:default=9999
	GrpcServerPort *int `json:"grpcServerPort,omitempty"`
}

// NetworkNodeStatus defines the observed state of NetworkNode
type NetworkNodeStatus struct {
	nddv1.ConditionedStatus `json:",inline"`
	ControllerRef           nddv1.Reference `json:"controllerRef,omitempty"`
	DeviceStatus            `json:",inline"`
}

// +kubebuilder:object:root=true
// +genclient
// +genclient:nonNamespaced

// NetworkNode is the Schema for the networknodes API
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="HEALTHY",type="string",JSONPath=".status.conditions[?(@.kind=='DeviceDriverHealthy')].status"
// +kubebuilder:printcolumn:name="CONFIGURED",type="string",JSONPath=".status.conditions[?(@.kind=='DeviceDriverConfigured')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.kind=='DeviceDriverReady')].status"
// +kubebuilder:printcolumn:name="ADDRESS",type="string",JSONPath=".spec.target.address",description="address to connect to the device'"
// +kubebuilder:printcolumn:name="CONN-KIND",type="string",JSONPath=".spec.deviceDriverKind",description="Kind of communication type to the device"
// +kubebuilder:printcolumn:name="TYPE",type="string",JSONPath=".status.deviceDetails.type",description="Type of device"
// +kubebuilder:printcolumn:name="KIND",type="string",JSONPath=".status.deviceDetails.kind",description="Kind of device"
// +kubebuilder:printcolumn:name="SWVERSION",type="string",JSONPath=".status.deviceDetails.swVersion",description="SW version of the device"
// +kubebuilder:printcolumn:name="MACADDRESS",type="string",JSONPath=".status.deviceDetails.macAddress",description="macAddress of the device"
// +kubebuilder:printcolumn:name="SERIALNBR",type="string",JSONPath=".status.deviceDetails.serialNumber",description="serialNumber of the device"
// +kubebuilder:printcolumn:name="GRPCSERVERPORT",type="string",JSONPath=".spec.grpcServerPort",description="grpc server port to connect to the devic driver"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={ndd,dvr},shortName=nn
type NetworkNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkNodeSpec   `json:"spec,omitempty"`
	Status NetworkNodeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NetworkNodeList contains a list of NetworkNode
type NetworkNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkNode `json:"items"`
}

// +kubebuilder:object:root=true

// A NetworkNodeUsage indicates that a resource is using a NetworkNode.
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="CONFIG-NAME",type="string",JSONPath=".NetworkNodeRef.name"
// +kubebuilder:printcolumn:name="RESOURCE-KIND",type="string",JSONPath=".resourceRef.kind"
// +kubebuilder:printcolumn:name="RESOURCE-NAME",type="string",JSONPath=".resourceRef.name"
// +kubebuilder:resource:scope=Cluster,categories={ndd,dvr},shortName=nnu
type NetworkNodeUsage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	nddv1.NetworkNodeUsage `json:",inline"`
}

// +kubebuilder:object:root=true

// NetworkNodeUsageList contains a list of NetworkNodeUsage
type NetworkNodeUsageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkNodeUsage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NetworkNode{}, &NetworkNodeList{})
	SchemeBuilder.Register(&NetworkNodeUsage{}, &NetworkNodeUsageList{})
}
