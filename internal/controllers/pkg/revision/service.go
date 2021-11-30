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

package revision

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pkgmetav1 "github.com/yndd/ndd-core/apis/pkg/meta/v1"
	v1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/yndd/ndd-runtime/pkg/meta"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func buildIntentService(intent *pkgmetav1.Intent, revision v1.PackageRevision, namespace string) *corev1.Service { // nolint:interfacer,gocyclo
	gnmiLabelName := strings.Join([]string{pkgmetav1.PrefixGnmiService, strings.Split(revision.GetName(), "-")[len(strings.Split(revision.GetName(), "-"))-1]}, "-")
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gnmiLabelName,
			Namespace: namespace,
			Labels: map[string]string{
				pkgmetav1.LabelPkgMeta: intent.GetName(),
			},
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(revision, v1.ProviderRevisionGroupVersionKind))},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				//pkgmetav1.LabelPkgMeta: intent.GetName(),
				"pkg.ndd.yndd.io/revision": revision.GetName(),
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "gnmi",
					Port:       pkgmetav1.GnmiServerPort,
					TargetPort: intstr.FromInt(pkgmetav1.GnmiServerPort),
					Protocol:   "TCP",
				},
			},
		},
	}
}

func buildIntentMetricService(intent *pkgmetav1.Intent, revision v1.PackageRevision, namespace string) *corev1.Service { // nolint:interfacer,gocyclo
	metricLabelName := strings.Join([]string{pkgmetav1.PrefixMetricService, strings.Split(revision.GetName(), "-")[len(strings.Split(revision.GetName(), "-"))-1]}, "-")
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      metricLabelName,
			Namespace: namespace,
			Labels: map[string]string{
				pkgmetav1.LabelPkgMeta: metricLabelName,
			},
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(revision, v1.ProviderRevisionGroupVersionKind))},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				pkgmetav1.LabelPkgMeta: metricLabelName,
				//"pkg.ndd.yndd.io/revision": revision.GetName(),
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "metrics",
					Port:       pkgmetav1.MetricServerPort,
					TargetPort: intstr.FromString("https"),
					Protocol:   "TCP",
				},
			},
		},
	}
}
