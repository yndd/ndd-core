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

// A PackageType is a type of package.
type PackageType string

// Types of packages.
const (
	ProviderPackageType PackageType = "Provider"
)

// MetaSpec are fields that every meta package type must implement.
type MetaSpec struct {
	// Semantic version constraints of Ndd that package is compatible with.
	Ndd *NddConstraints `json:"ndd,omitempty"`

	// Dependencies on other packages.
	DependsOn []Dependency `json:"dependsOn,omitempty"`
}

// NddConstraints specifies a packages compatibility with ndd versions.
type NddConstraints struct {
	// Semantic version constraints of ndd that package is compatible with.
	Version string `json:"version"`
}

// Dependency is a dependency on another package. One of Provider or Configuration may be supplied.
type Dependency struct {
	// Provider is the name of a Provider package image.
	//Provider *string `json:"provider,omitempty"`

	// Version is the semantic version constraints of the dependency image.
	//Version string `json:"version"`

	// Package is the OCI image name without a tag or digest.
	Package string `json:"package"`

	// Type is the type of package. Can be either Configuration or Provider.
	Type PackageType `json:"type"`

	// Constraints is a valid semver range, which will be used to select a valid
	// dependency version.
	Constraints string `json:"constraints"`
}
