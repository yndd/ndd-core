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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	pkgmetav1 "github.com/yndd/ndd-core/apis/pkg/meta/v1"
	pkgv1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/yndd/ndd-runtime/pkg/meta"
	corev1 "k8s.io/api/core/v1"
)

func renderService(cc *pkgmetav1.Provider, podSpec *pkgmetav1.PodSpec, c *pkgmetav1.ContainerSpec, extra *pkgmetav1.Extras, pr pkgv1.PackageRevision) *corev1.Service { // nolint:interfacer,gocyclo
	// we use the parent label key to get a consistent name for service discovery
	serviceName := getServiceName(pr.GetLabels()[pkgv1.ParentLabelKey], c.Container.Name, extra.Name)
	servicePrName := getServiceName(pr.GetName(), c.Container.Name, extra.Name)

	port := int32(443)
	if extra.Port != 0 {
		port = int32(extra.Port)
	}
	protocol := corev1.Protocol("TCP")
	if extra.Protocol != "" {
		protocol = corev1.Protocol(extra.Protocol)
	}
	targetPort := int(8443)
	if extra.TargetPort != 0 {
		targetPort = int(extra.TargetPort)
	}

	spec := corev1.ServiceSpec{
		Selector: map[string]string{
			getLabelKey(extra.Name): servicePrName,
		},
		Ports: []corev1.ServicePort{
			{
				Name:       extra.Name,
				Port:       port,
				TargetPort: intstr.FromInt(targetPort),
				Protocol:   protocol,
			},
		},
	}
	if strings.Contains(extra.Name, "headless") {
		spec.ClusterIP = "None"
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName, // this is the name w/o the revision
			Namespace: cc.Namespace,
			Labels: map[string]string{
				getLabelKey(extra.Name): servicePrName,
			},
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(pr, pkgv1.ProviderRevisionGroupVersionKind))},
		},
		Spec: spec,
	}
}

/*
const (
	metricsKey  = "metrics"
	profilerKey = "profiler"
	revisionKey = "revision"
	//packageNamespace = "pkg.ndd.yndd.io"
)

func buildProviderMetricServiceHTTPS(provider *pkgmetav1.Provider, revision v1.PackageRevision, namespace string) *corev1.Service { // nolint:interfacer,gocyclo
	metricHTTPSServiceName := strings.Join([]string{provider.GetName(), metricsKey, serviceSuffix}, "-")
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      metricHTTPSServiceName,
			Namespace: namespace,
			Labels: map[string]string{
				strings.Join([]string{pkgv1.PackageNamespace, metricsKey}, "/"): metricHTTPSServiceName,
			},
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(revision, v1.ProviderRevisionGroupVersionKind))},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				metricsKey: metricHTTPSServiceName,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       metricsKey,
					Port:       pkgmetav1.MetricServerPortHttps,
					TargetPort: intstr.FromString("https"),
					Protocol:   "TCP",
				},
			},
		},
	}
}

func buildProviderProfileService(provider *pkgmetav1.Provider, revision v1.PackageRevision, namespace string) *corev1.Service { // nolint:interfacer,gocyclo
	profileServiceName := strings.Join([]string{provider.GetName(), profilerKey, serviceSuffix}, "-")
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      profileServiceName,
			Namespace: namespace,
			Labels: map[string]string{
				strings.Join([]string{pkgv1.PackageNamespace, profilerKey}, "/"): profileServiceName,
			},
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(revision, v1.ProviderRevisionGroupVersionKind))},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				profilerKey: profileServiceName,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       profilerKey,
					Port:       8000,
					TargetPort: intstr.FromInt(8000),
					Protocol:   "TCP",
				},
			},
		},
	}
}
*/
