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
)

const (
	ConfigmapJsonConfig      = "config.json"
	LabelApplication         = "app"
	LabelNetworkDeviceDriver = "ndd"
	PrefixNetworkNode        = "ndd"
	PrefixConfigmap          = "ndd-cm"
	PrefixDeployment         = "ndd-dep"
	PrefixService            = "ndd-svc"
	Namespace                = "ndd-system"
	NamespaceLocalK8sDNS     = Namespace + "." + "svc.cluster.local:"
)

// DeviceDriverKind represents the kinds of device drivers are supported
// by the network device driver
type DeviceDriverKind string

const (
	// DeviceDriverKindGnmi operates using the gnmi specification
	DeviceDriverKindGnmi DeviceDriverKind = "gnmi"

	// DeviceDriverKindNetconf operates using the netconf specification
	DeviceDriverKindNetconf DeviceDriverKind = "netconf"
)

// TargetDetails contains the information necessary to communicate with
// the network node.
type TargetDetails struct {
	// Address holds the IP:port for accessing the network node
	// +kubebuilder:validation:Required
	Address *string `json:"address"`

	// Proxy used to communicate to the target network node
	// +kubebuilder:validation:Optional
	Proxy *string `json:"proxy,omitempty"`

	// The name of the secret containing the credentials (requires
	// keys "username" and "password").
	// +kubebuilder:validation:Required
	CredentialsName *string `json:"credentialsName"`

	// The name of the secret containing the credentials (requires
	// keys "TLSCA" and "TLSCert", " TLSKey").
	// +kubebuilder:validation:Optional
	TLSCredentialsName *string `json:"tlsCredentialsName,omitempty"`

	// SkipVerify disables verification of server certificates when using
	// HTTPS to connect to the Target. This is required when the server
	// certificate is self-signed, but is insecure because it allows a
	// man-in-the-middle to intercept the connection.
	// +kubebuilder:default:=false
	// +kubebuilder:validation:Optional
	SkipVerify *bool `json:"skpVerify,omitempty"`

	// Insecure runs the communication in an insecure manner
	// +kubebuilder:default:=false
	// +kubebuilder:validation:Optional
	Insecure *bool `json:"insecure,omitempty"`

	// Encoding defines the gnmi encoding
	// +kubebuilder:validation:Enum=`JSON`;`BYTES`;`PROTO`;`ASCII`;`JSON_IETF`
	// +kubebuilder:default=JSON_IETF
	// +kubebuilder:validation:Optional
	Encoding *string `json:"encoding,omitempty"`
}

/*
// DeviceDriverDetails defines the device driver details to connect to the network node
type DeviceDriverDetails struct {
	// Kind defines the device driver kind
	// +kubebuilder:default:=gnmi
	// +kubebuilder:validation:Required
	Kind *DeviceDriverKind `json:"kind"`
}
*/

/*
type GrpcServerDetails struct {
	// Port defines the port of the GRPC server for the device driver
	// +kubebuilder:default:=9999
	// +kubebuilder:validation:Required
	Port *int `json:"port"`
}
*/

// DeviceDetails collects information about the deiscovered device
type DeviceDetails struct {
	// the Type of device the device driver is connected to
	Type *nddv1.DeviceType `json:"type,omitempty"`

	// Host name of the device the device driver is connected to
	HostName *string `json:"hostname,omitempty"`

	// the Kind of device the device driver is connected to
	Kind *string `json:"kind,omitempty"`

	// SW version that is running on the device
	SwVersion *string `json:"swVersion,omitempty"`

	// the Mac address of the device the device driver is connected to
	MacAddress *string `json:"macAddress,omitempty"`

	// the Serial Number of the device the device driver is connected to
	SerialNumber *string `json:"serialNumber,omitempty"`

	// Supported Encodings by the device
	SupportedEncodings []string `json:"supportedEncodings,omitempty"`
}

// DeviceStatus defines the observed state of the Device
type DeviceStatus struct {
	// The discovered DeviceDetails
	DeviceDetails *DeviceDetails `json:"deviceDetails,omitempty"`

	// UsedNetworkNodeSpec identifies the used networkNode spec when installed
	UsedNetworkNodeSpec *NetworkNodeSpec `json:"usedNetworkNodeSpec,omitempty"`

	// UsedDeviceDriverSpec identifies the used deviceDriver spec when installed
	UsedDeviceDriverSpec *DeviceDriverSpec `json:"usedDeviceDriverSpec,omitempty"`
}
