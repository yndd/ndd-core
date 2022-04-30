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
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pkgmetav1 "github.com/yndd/ndd-core/apis/pkg/meta/v1"
	v1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/yndd/ndd-runtime/pkg/meta"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	serviceSuffix = "svc"
	certSuffix    = "serving-cert"

	metricsKey       = "metrics"
	profilerKey      = "profiler"
	revisionKey      = "revision"
	packageNamespace = "pkg.ndd.yndd.io"
)

func buildIntentService(intent *pkgmetav1.Intent, revision v1.PackageRevision, namespace string) *corev1.Service { // nolint:interfacer,gocyclo
	//gnmiLabelName := strings.Join([]string{pkgmetav1.PrefixGnmiService, strings.Split(revision.GetName(), "-")[len(strings.Split(revision.GetName(), "-"))-1]}, "-")
	gnmiLabelName := strings.Join([]string{pkgmetav1.PrefixGnmiService, revision.GetName()}, "-")
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gnmiLabelName,
			Namespace: namespace,
			Labels: map[string]string{
				pkgmetav1.LabelPkgMeta: intent.GetName(),
			},
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(revision, v1.IntentRevisionGroupVersionKind))},
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

func buildIntentMetricServiceHTTPS(intent *pkgmetav1.Intent, revision v1.PackageRevision, namespace string) *corev1.Service { // nolint:interfacer,gocyclo
	//metricLabelName := strings.Join([]string{pkgmetav1.PrefixMetricService, strings.Split(revision.GetName(), "-")[len(strings.Split(revision.GetName(), "-"))-1]}, "-")
	metricLabelName := strings.Join([]string{pkgmetav1.PrefixMetricService, revision.GetName(), "https"}, "-")
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      metricLabelName,
			Namespace: namespace,
			/*
				Annotations: map[string]string{
					"prometheus.io/path":   "/metrics",
					"prometheus.io/scheme": "https",
					"prometheus.io/insecure_skip_verify": "true",
					"prometheus.io/port":   strconv.Itoa(pkgmetav1.MetricServerPortHttps),
					"prometheus.io/scrape": "true",
				},
			*/
			Labels: map[string]string{
				pkgmetav1.LabelPkgMeta: metricLabelName,
			},
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(revision, v1.IntentRevisionGroupVersionKind))},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				pkgmetav1.LabelPkgMeta: metricLabelName,
				//"pkg.ndd.yndd.io/revision": revision.GetName(),
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "metrics",
					Port:       pkgmetav1.MetricServerPortHttps,
					TargetPort: intstr.FromString("https"),
					Protocol:   "TCP",
				},
			},
		},
	}
}

func buildIntentMetricServiceHTTP(intent *pkgmetav1.Intent, revision v1.PackageRevision, namespace string) *corev1.Service { // nolint:interfacer,gocyclo
	//metricLabelName := strings.Join([]string{pkgmetav1.PrefixMetricService, strings.Split(revision.GetName(), "-")[len(strings.Split(revision.GetName(), "-"))-1]}, "-")
	metricLabelNameHttp := strings.Join([]string{pkgmetav1.PrefixMetricService, revision.GetName(), "http"}, "-")
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      metricLabelNameHttp,
			Namespace: namespace,
			Annotations: map[string]string{
				"prometheus.io/path":   "/metrics",
				"prometheus.io/port":   strconv.Itoa(pkgmetav1.MetricServerPortHttp),
				"prometheus.io/scrape": "true",
			},
			Labels: map[string]string{
				pkgmetav1.LabelHttpPkgMeta: metricLabelNameHttp,
			},
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(revision, v1.IntentRevisionGroupVersionKind))},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				pkgmetav1.LabelHttpPkgMeta: metricLabelNameHttp,
				//"pkg.ndd.yndd.io/revision": revision.GetName(),
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "metrics",
					Port:       pkgmetav1.MetricServerPortHttp,
					TargetPort: intstr.FromInt(pkgmetav1.MetricServerPortHttp),
					Protocol:   "TCP",
				},
			},
		},
	}
}

/*
func buildProviderService(provider *pkgmetav1.Provider, revision v1.PackageRevision, namespace string) *corev1.Service { // nolint:interfacer,gocyclo
	gnmiServiceName := strings.Join([]string{revision.GetName(), "gnmi", "svc"}, "-")
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gnmiServiceName,
			Namespace: namespace,
			Labels: map[string]string{
				"gnmi": gnmiServiceName,
			},
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(revision, v1.ProviderRevisionGroupVersionKind))},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				//pkgmetav1.LabelPkgMeta: intent.GetName(),
				"gnmi": gnmiServiceName,
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
*/

func buildProviderMetricServiceHTTPS(provider *pkgmetav1.Provider, revision v1.PackageRevision, namespace string) *corev1.Service { // nolint:interfacer,gocyclo
	metricHTTPSServiceName := strings.Join([]string{provider.GetName(), metricsKey, serviceSuffix}, "-")
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      metricHTTPSServiceName,
			Namespace: namespace,
			Labels: map[string]string{
				strings.Join([]string{packageNamespace, metricsKey}, "/"): metricHTTPSServiceName,
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
				strings.Join([]string{packageNamespace, profilerKey}, "/"): profileServiceName,
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

/*
func buildProviderWebhookService(provider *pkgmetav1.Provider, revision v1.PackageRevision, namespace string) *corev1.Service { // nolint:interfacer,gocyclo
	//metricLabelName := strings.Join([]string{pkgmetav1.PrefixMetricService, strings.Split(revision.GetName(), "-")[len(strings.Split(revision.GetName(), "-"))-1]}, "-")
	webhookServiceName := strings.Join([]string{revision.GetName(), "webhook", "svc"}, "-")
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      webhookServiceName,
			Namespace: namespace,
			Labels: map[string]string{
				"webhook": webhookServiceName,
			},
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(revision, v1.ProviderRevisionGroupVersionKind))},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"webhook": webhookServiceName,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "webhook",
					Port:       443,
					TargetPort: intstr.FromInt(9443),
					Protocol:   "TCP",
				},
			},
		},
	}
}
*/
