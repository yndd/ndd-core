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
)

// PackageSpec defines the desired state of Package
type PackageSpec struct {
	// Package is the name of the package that is being requested.
	Package string `json:"package"`

	// AutoPilot specifies how the provider operates
	// When set to true the provider applies delta/diff changes to the device
	// manged resources automatically, if set to false the provider will report
	// the delta and the operator should intervene what to do with the delta/diffs
	// Defaults to true. Can be disabled by explicitly setting to flase.
	// +optional
	// +kubebuilder:default=true
	AutoPilot *bool `json:"autoPilot,omitempty"`

	// RevisionActivationPolicy specifies how the package controller should
	// update from one revision to the next. Options are Automatic or Manual.
	// Default is Automatic.
	// +optional
	// +kubebuilder:default=Automatic
	RevisionActivationPolicy *RevisionActivationPolicy `json:"revisionActivationPolicy,omitempty"`

	// RevisionHistoryLimit dictates how the package controller cleans up old
	// inactive package revisions.
	// Defaults to 1. Can be disabled by explicitly setting to 0.
	// +optional
	// +kubebuilder:default=1
	RevisionHistoryLimit *int64 `json:"revisionHistoryLimit,omitempty"`

	// PackagePullSecrets are named secrets in the same namespace that can be used
	// to fetch packages from private registries.
	// +optional
	PackagePullSecrets []corev1.LocalObjectReference `json:"packagePullSecrets,omitempty"`

	// PackagePullPolicy defines the pull policy for the package.
	// Default is IfNotPresent.
	// +optional
	// +kubebuilder:default=IfNotPresent
	PackagePullPolicy *corev1.PullPolicy `json:"packagePullPolicy,omitempty"`

	// SkipDependencyResolution indicates to the package manager whether to skip
	// resolving dependencies for a package. Setting this value to true may have
	// unintended consequences.
	// Default is false.
	// +optional
	// +kubebuilder:default=false
	SkipDependencyResolution *bool `json:"skipDependencyResolution,omitempty"`
}

// PackageStatus defines the observed state of Package
type PackageStatus struct {
	// CurrentRevision is the name of the current package revision. It will
	// reflect the most up to date revision, whether it has been activated or
	// not.
	CurrentRevision string `json:"currentRevision,omitempty"`

	// CurrentIdentifier is the most recent package source that was used to
	// produce a revision. The package manager uses this field to determine
	// whether to check for package updates for a given source when
	// packagePullPolicy is set to IfNotPresent. Manually removing this field
	// will cause the package manager to check that the current revision is
	// correct for the given package source.
	CurrentIdentifier string `json:"currentIdentifier,omitempty"`
}
