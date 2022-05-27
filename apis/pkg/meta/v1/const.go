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

const (
	GnmiServerPort        = 9999
	MetricServerPortHttp  = 9997
	MetricServerPortHttps = 8443
	PrefixGnmiService     = "nddo-gnmi-svc"
	PrefixMetricService   = "nddo-metrics-svc"
	Namespace             = "ndd-system"
	NamespaceLocalK8sDNS  = Namespace + "." + "svc.cluster.local:"
	LabelPkgMeta          = "app"
	LabelHttpPkgMeta      = "app-http"
	VendorTypeLabelKey    = Group + "/" + "vendorType"
)
