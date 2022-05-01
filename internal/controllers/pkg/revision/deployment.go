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
	"fmt"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pkgmetav1 "github.com/yndd/ndd-core/apis/pkg/meta/v1"
	pkgv1 "github.com/yndd/ndd-core/apis/pkg/v1"
	v1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/yndd/ndd-runtime/pkg/meta"
	"github.com/yndd/ndd-runtime/pkg/utils"
)

const (
	containerStartupCmd = "/manager"

	userGroup = 2000
)

func getPullPolicy(revision v1.PackageRevision) corev1.PullPolicy {
	pullPolicy := corev1.PullIfNotPresent
	if revision.GetPackagePullPolicy() != nil {
		pullPolicy = *revision.GetPackagePullPolicy()
	}
	return pullPolicy
}

func getEnv() []corev1.EnvVar {
	// environment parameters used in the deployment/statefulset
	envNameSpace := corev1.EnvVar{
		Name: "POD_NAMESPACE",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.namespace",
			},
		},
	}
	envPodIP := corev1.EnvVar{
		Name: "POD_IP",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.podIP",
			},
		},
	}
	envPodName := corev1.EnvVar{
		Name: "POD_NAME",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.name",
			},
		},
	}
	return []corev1.EnvVar{
		envNameSpace,
		envPodIP,
		envPodName,
	}
}

func getProxyArgs() []string {
	return []string{
		"--secure-listen-address=0.0.0.0:8443",
		"--upstream=http://127.0.0.1:8080/",
		"--logtostderr=true",
		"--v=10",
	}
}

func getArgs(provider *pkgmetav1.Provider, revision v1.PackageRevision) []string {
	args := []string{
		"start",
		"--debug",
		fmt.Sprintf("--revision=%s", revision.GetName()),
		fmt.Sprintf("--revision-namespace=%s", revision.GetNamespace()),
		fmt.Sprintf("--controller-config-name=%s", provider.GetName()),
		fmt.Sprintf("--autopilot=%s", strconv.FormatBool(revision.GetAutoPilot())),
	}
	return args
}

func getContainers(provider *pkgmetav1.Provider, revision v1.PackageRevision, cc *v1.ControllerConfig, namespace string) []corev1.Container {
	containers := []corev1.Container{}
	containers = append(containers, getKubeProxyContainer())
	containers = append(containers, getControllerContainer(provider, revision))

	return containers
}

func getKubeProxyContainer() corev1.Container {
	return corev1.Container{
		Name:  "kube-rbac-proxy",
		Image: "gcr.io/kubebuilder/kube-rbac-proxy:v0.8.0",
		Args:  getProxyArgs(),
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: 8443,
				Name:          "https",
			},
		},
	}
}

func getControllerContainer(provider *pkgmetav1.Provider, revision v1.PackageRevision) corev1.Container {
	return corev1.Container{
		Name:            "controller",
		Image:           provider.Spec.Controller.Image,
		ImagePullPolicy: getPullPolicy(revision),
		SecurityContext: getSecurityContext(),
		Args:            getArgs(provider, revision),
		Env:             getEnv(),
		Command: []string{
			containerStartupCmd,
		},
		VolumeMounts: getVolumeMounts(),
	}
}

func getPodSecurityContext() *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{
		RunAsUser:    utils.Int64Ptr(userGroup),
		RunAsGroup:   utils.Int64Ptr(userGroup),
		RunAsNonRoot: utils.BoolPtr(true),
	}
}

func getSecurityContext() *corev1.SecurityContext {
	return &corev1.SecurityContext{
		RunAsUser:                utils.Int64Ptr(userGroup),
		RunAsGroup:               utils.Int64Ptr(userGroup),
		AllowPrivilegeEscalation: utils.BoolPtr(false),
		Privileged:               utils.BoolPtr(false),
		RunAsNonRoot:             utils.BoolPtr(true),
	}
}

func getVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      profilerKey,
			MountPath: fmt.Sprintf("/%s", profilerKey),
		},
	}

}

func getVolumes() []corev1.Volume {
	return []corev1.Volume{
		{
			Name: profilerKey,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
}

func buildProviderDeployment(provider *pkgmetav1.Provider, revision v1.PackageRevision, cc *v1.ControllerConfig, namespace string) (*corev1.ServiceAccount, *appsv1.Deployment) { // nolint:interfacer,gocyclo
	s := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:            revision.GetName(),
			Namespace:       namespace,
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(revision, v1.ProviderRevisionGroupVersionKind))},
		},
	}

	profileServiceName := strings.Join([]string{provider.GetName(), profilerKey, serviceSuffix}, "-")
	metricHTTPSServiceName := strings.Join([]string{provider.GetName(), metricsKey, serviceSuffix}, "-")
	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            revision.GetName(),
			Namespace:       namespace,
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(revision, v1.ProviderRevisionGroupVersionKind))},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: utils.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					strings.Join([]string{pkgv1.PackageNamespace, revisionKey}, "/"): revision.GetName(),
					strings.Join([]string{pkgv1.PackageNamespace, metricsKey}, "/"):  metricHTTPSServiceName,
					strings.Join([]string{pkgv1.PackageNamespace, profilerKey}, "/"): profileServiceName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      provider.GetName(),
					Namespace: namespace,
					Labels: map[string]string{
						strings.Join([]string{pkgv1.PackageNamespace, revisionKey}, "/"): revision.GetName(),
						strings.Join([]string{pkgv1.PackageNamespace, metricsKey}, "/"):  metricHTTPSServiceName,
						strings.Join([]string{pkgv1.PackageNamespace, profilerKey}, "/"): profileServiceName,
					},
				},
				Spec: corev1.PodSpec{
					SecurityContext:    getPodSecurityContext(),
					ServiceAccountName: s.GetName(),
					ImagePullSecrets:   revision.GetPackagePullSecrets(),
					Containers:         getContainers(provider, revision, cc, namespace),
					Volumes:            getVolumes(),
				},
			},
		},
	}
	if cc != nil {
		s.Labels = cc.Labels
		s.Annotations = cc.Annotations
		d.Labels = cc.Labels
		d.Annotations = cc.Annotations
		if cc.Spec.Metadata != nil {
			d.Spec.Template.Annotations = cc.Spec.Metadata.Annotations
		}
		if cc.Spec.Replicas != nil {
			d.Spec.Replicas = cc.Spec.Replicas
		}
		if cc.Spec.Image != nil {
			d.Spec.Template.Spec.Containers[0].Image = *cc.Spec.Image
		}
		if len(cc.Spec.Ports) > 0 {
			d.Spec.Template.Spec.Containers[0].Ports = cc.Spec.Ports
		}
		if cc.Spec.NodeSelector != nil {
			d.Spec.Template.Spec.NodeSelector = cc.Spec.NodeSelector
		}
		if cc.Spec.ServiceAccountName != nil {
			d.Spec.Template.Spec.ServiceAccountName = *cc.Spec.ServiceAccountName
		}
		if cc.Spec.NodeName != nil {
			d.Spec.Template.Spec.NodeName = *cc.Spec.NodeName
		}
		if cc.Spec.PodSecurityContext != nil {
			d.Spec.Template.Spec.SecurityContext = cc.Spec.PodSecurityContext
		}
		if cc.Spec.SecurityContext != nil {
			d.Spec.Template.Spec.Containers[0].SecurityContext = cc.Spec.SecurityContext
		}
		if len(cc.Spec.ImagePullSecrets) > 0 {
			d.Spec.Template.Spec.ImagePullSecrets = cc.Spec.ImagePullSecrets
		}
		if cc.Spec.Affinity != nil {
			d.Spec.Template.Spec.Affinity = cc.Spec.Affinity
		}
		if len(cc.Spec.Tolerations) > 0 {
			d.Spec.Template.Spec.Tolerations = cc.Spec.Tolerations
		}
		if cc.Spec.PriorityClassName != nil {
			d.Spec.Template.Spec.PriorityClassName = *cc.Spec.PriorityClassName
		}
		if cc.Spec.RuntimeClassName != nil {
			d.Spec.Template.Spec.RuntimeClassName = cc.Spec.RuntimeClassName
		}
		if cc.Spec.ResourceRequirements != nil {
			d.Spec.Template.Spec.Containers[0].Resources = *cc.Spec.ResourceRequirements
		}
		if len(cc.Spec.Args) > 0 {
			d.Spec.Template.Spec.Containers[0].Args = cc.Spec.Args
		}
		if len(cc.Spec.EnvFrom) > 0 {
			d.Spec.Template.Spec.Containers[0].EnvFrom = cc.Spec.EnvFrom
		}
		if len(cc.Spec.Env) > 0 {
			d.Spec.Template.Spec.Containers[0].Env = cc.Spec.Env
		}
	}
	return s, d
}

func buildIntentDeployment(intent *pkgmetav1.Intent, revision v1.PackageRevision, cc *v1.ControllerConfig, namespace string) (*corev1.ServiceAccount, *appsv1.Deployment) { // nolint:interfacer,gocyclo
	//metricLabelName := strings.Join([]string{pkgmetav1.PrefixServiceMetric, strings.Split(intent.GetName(), "-")[len(strings.Split(intent.GetName(), "-"))-1]}, "-")
	//metricLabelName := strings.Join([]string{pkgmetav1.PrefixMetricService, strings.Split(revision.GetName(), "-")[len(strings.Split(revision.GetName(), "-"))-1]}, "-")
	metricLabelNameHttp := strings.Join([]string{pkgmetav1.PrefixMetricService, revision.GetName(), "http"}, "-")
	metricLabelNameHttps := strings.Join([]string{pkgmetav1.PrefixMetricService, revision.GetName(), "https"}, "-")
	s := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:            revision.GetName(),
			Namespace:       namespace,
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(revision, v1.IntentRevisionGroupVersionKind))},
		},
	}
	pullPolicy := corev1.PullIfNotPresent
	if revision.GetPackagePullPolicy() != nil {
		pullPolicy = *revision.GetPackagePullPolicy()
	}
	// environment parameters used in the deployment
	envNameSpace := corev1.EnvVar{
		Name: "POD_NAMESPACE",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.namespace",
			},
		},
	}
	envPodIP := corev1.EnvVar{
		Name: "POD_IP",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.podIP",
			},
		},
	}
	envPodName := corev1.EnvVar{
		Name: "POD_NAME",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.name",
			},
		},
	}

	args := []string{
		"start",
		"--debug",
	}

	argsProxy := []string{
		"--secure-listen-address=0.0.0.0:8443",
		"--upstream=http://127.0.0.1:9997/",
		"--logtostderr=true",
		"--v=10",
	}

	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            revision.GetName(),
			Namespace:       namespace,
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(revision, v1.IntentRevisionGroupVersionKind))},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: utils.Int32Ptr(0),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"pkg.ndd.yndd.io/revision": revision.GetName(),
					pkgmetav1.LabelPkgMeta:     metricLabelNameHttps,
					pkgmetav1.LabelHttpPkgMeta: metricLabelNameHttp,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      intent.GetName(),
					Namespace: namespace,
					Labels: map[string]string{
						"pkg.ndd.yndd.io/revision": revision.GetName(),
						pkgmetav1.LabelPkgMeta:     metricLabelNameHttps,
						pkgmetav1.LabelHttpPkgMeta: metricLabelNameHttp,
					},
				},
				Spec: corev1.PodSpec{
					SecurityContext:    getPodSecurityContext(),
					ServiceAccountName: s.GetName(),
					ImagePullSecrets:   revision.GetPackagePullSecrets(),
					Containers: []corev1.Container{
						{
							Name:  "kube-rbac-proxy",
							Image: "gcr.io/kubebuilder/kube-rbac-proxy:v0.8.0",
							Args:  argsProxy,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8443,
									Name:          "https",
								},
							},
						},
						{
							Name:            "intent",
							Image:           intent.Spec.Controller.Image,
							ImagePullPolicy: pullPolicy,
							SecurityContext: getSecurityContext(),
							Args:            args,
							Env: []corev1.EnvVar{
								envNameSpace,
								envPodIP,
								envPodName,
							},
							Command: []string{
								"/manager",
							},
						},
					},
				},
			},
		},
	}
	if cc != nil {
		s.Labels = cc.Labels
		s.Annotations = cc.Annotations
		d.Labels = cc.Labels
		d.Annotations = cc.Annotations
		if cc.Spec.Metadata != nil {
			d.Spec.Template.Annotations = cc.Spec.Metadata.Annotations
		}
		if cc.Spec.Replicas != nil {
			d.Spec.Replicas = cc.Spec.Replicas
		}
		if cc.Spec.Image != nil {
			d.Spec.Template.Spec.Containers[0].Image = *cc.Spec.Image
		}
		if len(cc.Spec.Ports) > 0 {
			d.Spec.Template.Spec.Containers[0].Ports = cc.Spec.Ports
		}
		if cc.Spec.NodeSelector != nil {
			d.Spec.Template.Spec.NodeSelector = cc.Spec.NodeSelector
		}
		if cc.Spec.ServiceAccountName != nil {
			d.Spec.Template.Spec.ServiceAccountName = *cc.Spec.ServiceAccountName
		}
		if cc.Spec.NodeName != nil {
			d.Spec.Template.Spec.NodeName = *cc.Spec.NodeName
		}
		if cc.Spec.PodSecurityContext != nil {
			d.Spec.Template.Spec.SecurityContext = cc.Spec.PodSecurityContext
		}
		if cc.Spec.SecurityContext != nil {
			d.Spec.Template.Spec.Containers[0].SecurityContext = cc.Spec.SecurityContext
		}
		if len(cc.Spec.ImagePullSecrets) > 0 {
			d.Spec.Template.Spec.ImagePullSecrets = cc.Spec.ImagePullSecrets
		}
		if cc.Spec.Affinity != nil {
			d.Spec.Template.Spec.Affinity = cc.Spec.Affinity
		}
		if len(cc.Spec.Tolerations) > 0 {
			d.Spec.Template.Spec.Tolerations = cc.Spec.Tolerations
		}
		if cc.Spec.PriorityClassName != nil {
			d.Spec.Template.Spec.PriorityClassName = *cc.Spec.PriorityClassName
		}
		if cc.Spec.RuntimeClassName != nil {
			d.Spec.Template.Spec.RuntimeClassName = cc.Spec.RuntimeClassName
		}
		if cc.Spec.ResourceRequirements != nil {
			d.Spec.Template.Spec.Containers[0].Resources = *cc.Spec.ResourceRequirements
		}
		if len(cc.Spec.Args) > 0 {
			d.Spec.Template.Spec.Containers[0].Args = cc.Spec.Args
		}
		if len(cc.Spec.EnvFrom) > 0 {
			d.Spec.Template.Spec.Containers[0].EnvFrom = cc.Spec.EnvFrom
		}
		if len(cc.Spec.Env) > 0 {
			d.Spec.Template.Spec.Containers[0].Env = cc.Spec.Env
		}
	}
	return s, d
}
