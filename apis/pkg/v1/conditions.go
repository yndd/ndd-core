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

// Package v1 contains API Schema definitions for the pkg v1 API group
//+kubebuilder:object:generate=true
//+groupName=pkg.ndd.yndd.io
package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	nddv1 "github.com/yndd/ndd-runtime/apis/common/v1"
)

// Condition Kinds.
const (
	// A PackageInstalled indicates whether a package has been installed.
	ConditionKindPackageInstalled nddv1.ConditionKind = "PackageInstalled"

	// A PackageHealthy indicates whether a package is healthy.
	ConditionKindPackageHealthy nddv1.ConditionKind = "PackageHealthy"
)

// ConditionReasons a package is or is not installed.
const (
	ConditionReasonUnpacking     nddv1.ConditionReason = "UnpackingPackage"
	ConditionReasonInactive      nddv1.ConditionReason = "InactivePackageRevision"
	ConditionReasonActive        nddv1.ConditionReason = "ActivePackageRevision"
	ConditionReasonUnhealthy     nddv1.ConditionReason = "UnhealthyPackageRevision"
	ConditionReasonHealthy       nddv1.ConditionReason = "HealthyPackageRevision"
	ConditionReasonUnknownHealth nddv1.ConditionReason = "UnknownPackageRevisionHealth"
)

// Unpacking indicates that the package manager is waiting for a package
// revision to be unpacked.
func Unpacking() nddv1.Condition {
	return nddv1.Condition{
		Kind:               ConditionKindPackageInstalled,
		Status:             corev1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Reason:             ConditionReasonUnpacking,
	}
}

// Inactive indicates that the package manager is waiting for a package
// revision to be transitioned to an active state.
func Inactive() nddv1.Condition {
	return nddv1.Condition{
		Kind:               ConditionKindPackageInstalled,
		Status:             corev1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Reason:             ConditionReasonInactive,
	}
}

// Active indicates that the package manager has installed and activated
// a package revision.
func Active() nddv1.Condition {
	return nddv1.Condition{
		Kind:               ConditionKindPackageInstalled,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             ConditionReasonActive,
	}
}

// Unhealthy indicates that the current revision is unhealthy.
func Unhealthy() nddv1.Condition {
	return nddv1.Condition{
		Kind:               ConditionKindPackageHealthy,
		Status:             corev1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Reason:             ConditionReasonUnhealthy,
	}
}

// Healthy indicates that the current revision is healthy.
func Healthy() nddv1.Condition {
	return nddv1.Condition{
		Kind:               ConditionKindPackageHealthy,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             ConditionReasonHealthy,
	}
}

// UnknownHealth indicates that the health of the current revision is unknown.
func UnknownHealth() nddv1.Condition {
	return nddv1.Condition{
		Kind:               ConditionKindPackageHealthy,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             ConditionReasonUnknownHealth,
	}
}
