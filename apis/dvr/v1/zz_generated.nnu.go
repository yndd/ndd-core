// +build !ignore_autogenerated
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
// Code generated by ndd-gen. DO NOT EDIT.

package v1

import nddv1 "github.com/yndd/ndd-runtime/apis/common/v1"

// GetNetworkNodeReference of this NetworkNodeUsage.
func (p *NetworkNodeUsage) GetNetworkNodeReference() nddv1.Reference {
	return p.NetworkNodeReference
}

// GetResourceReference of this NetworkNodeUsage.
func (p *NetworkNodeUsage) GetResourceReference() nddv1.TypedReference {
	return p.ResourceReference
}

// SetNetworkNodeReference of this NetworkNodeUsage.
func (p *NetworkNodeUsage) SetNetworkNodeReference(r nddv1.Reference) {
	p.NetworkNodeReference = r
}

// SetResourceReference of this NetworkNodeUsage.
func (p *NetworkNodeUsage) SetResourceReference(r nddv1.TypedReference) {
	p.ResourceReference = r
}
