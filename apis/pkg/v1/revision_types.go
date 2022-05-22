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
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// PackageRevisionDesiredState is the desired state of the package revision.
type PackageRevisionDesiredState string

const (
	// PackageRevisionActive is an active package revision.
	PackageRevisionActive PackageRevisionDesiredState = "Active"

	// PackageRevisionInactive is an inactive package revision.
	PackageRevisionInactive PackageRevisionDesiredState = "Inactive"
)

// PackageRevisionSpec defines the desired state of Revision
type PackageRevisionSpec struct {
	// ControllerRef references a Controllerg resource that will be
	// used to configure the packaged controller Deployment.
	// +optional
	ControllerReference *nddv1.Reference `json:"controllerRef,omitempty"`

	// Kind is the kind of package
	// +kubebuilder:validation:Enum=`worker`;`reconciler`
	Kind Kind `json:"kind,omitempty"`

	// DesiredState of the PackageRevision. Can be either Active or Inactive.
	DesiredState PackageRevisionDesiredState `json:"desiredState"`

	// Package image used by install Pod to extract package contents.
	PackageImage string `json:"packageImage"`

	// PackagePullSecrets are named secrets in the same namespace that can be
	// used to fetch packages from private registries. They are also applied to
	// any images pulled for the package, such as a provider's controller image.
	// +optional
	PackagePullSecrets []corev1.LocalObjectReference `json:"packagePullSecrets,omitempty"`

	// PackagePullPolicy defines the pull policy for the package. It is also
	// applied to any images pulled for the package, such as a provider's
	// controller image.
	// Default is IfNotPresent.
	// +optional
	// +kubebuilder:default=IfNotPresent
	PackagePullPolicy *corev1.PullPolicy `json:"packagePullPolicy,omitempty"`

	// Revision number. Indicates when the revision will be garbage collected
	// based on the parent's RevisionHistoryLimit.
	Revision int64 `json:"revision"`

	// SkipDependencyResolution indicates to the package manager whether to skip
	// resolving dependencies for a package. Setting this value to true may have
	// unintended consequences.
	// Default is false.
	// +optional
	// +kubebuilder:default=false
	SkipDependencyResolution *bool `json:"skipDependencyResolution,omitempty"`
}

// PackageRevisionStatus defines the observed state of a PackageRevision
type PackageRevisionStatus struct {
	nddv1.ConditionedStatus `json:",inline"`
	ControllerRef           nddv1.Reference `json:"controllerRef,omitempty"`

	// References to objects owned by PackageRevision.
	ObjectRefs []nddv1.TypedReference `json:"objectRefs,omitempty"`

	// Dependency information.
	FoundDependencies     int64 `json:"foundDependencies,omitempty"`
	InstalledDependencies int64 `json:"installedDependencies,omitempty"`
	InvalidDependencies   int64 `json:"invalidDependencies,omitempty"`

	// PermissionRequests made by this package. The package declares that its
	// controller needs these permissions to run. The RBAC manager is
	// responsible for granting them.
	PermissionRequests []rbacv1.PolicyRule `json:"permissionRequests,omitempty"`
	// Api CRDs used by this package.
	//Apis []pkgmetav1.Api `json:"apis,omitempty"`
	// Pods used by this package in the Controller
	//Pods []pkgmetav1.PodSpec `json:"pods,omitempty"`
}
