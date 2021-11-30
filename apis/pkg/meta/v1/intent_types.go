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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	GnmiServerPort       = 9999
	MetricServerPort     = 8443
	PrefixGnmiService    = "nddo-gnmi-svc"
	PrefixMetricService  = "nddo-metrics-svc"
	Namespace            = "ndd-system"
	NamespaceLocalK8sDNS = Namespace + "." + "svc.cluster.local:"
	LabelPkgMeta         = "app"
)

// IntentSpec specifies the configuration of a Intent.
type IntentSpec struct {
	// Configuration for the packaged Intent's controller.
	Controller ControllerSpec `json:"controller"`

	MetaSpec `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

// A Intent is the description of a Ndd Intent package.
type Intent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec IntentSpec `json:"spec"`
}

// Hub marks this type as the conversion hub.
func (p *Intent) Hub() {}

func init() {
	SchemeBuilder.Register(&Intent{})
}
