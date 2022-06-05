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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pkgmetav1 "github.com/yndd/ndd-core/apis/pkg/meta/v1"
	pkgv1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/yndd/ndd-runtime/pkg/meta"
	"github.com/yndd/ndd-runtime/pkg/utils"
)

func renderProviderDeployment(pm *pkgmetav1.Provider, podSpec *pkgmetav1.PodSpec, pr pkgv1.PackageRevision, o *Options) *appsv1.Deployment {
	s := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            pr.GetName(),
			Namespace:       pm.Namespace,
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(pr, pkgv1.ProviderRevisionGroupVersionKind))},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: utils.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: getLabels(podSpec, pr),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pr.GetName(),
					Namespace: pm.Namespace,
					Labels:    getLabels(podSpec, pr),
				},
				Spec: corev1.PodSpec{
					Hostname:           pr.GetName(),
					SecurityContext:    getPodSecurityContext(),
					ServiceAccountName: renderServiceAccount(pm, podSpec, pr).GetName(),
					ImagePullSecrets:   pr.GetPackagePullSecrets(),
					Containers:         getContainers(pm, podSpec, pr.GetPackagePullPolicy(), o),
					Volumes:            getVolumes(podSpec, pr),
				},
			},
		},
	}

	return s
}

/*
func getPullPolicy(revision v1.PackageRevision) corev1.PullPolicy {
	pullPolicy := corev1.PullIfNotPresent
	if revision.GetPackagePullPolicy() != nil {
		pullPolicy = *revision.GetPackagePullPolicy()
	}
	return pullPolicy
}




func getControllerContainer(p *pkgmetav1.Provider, revision v1.PackageRevision, crdNames []string) corev1.Container {
	return corev1.Container{
		Name:            "controller",
		Image:           p.Spec.Pod.Image,
		ImagePullPolicy: getPullPolicy(revision),
		SecurityContext: getSecurityContext(),
		Args:            getArgs(p, revision, crdNames),
		Env:             getEnv(),
		Command: []string{
			containerStartupCmd,
		},
		VolumeMounts: getVolumeMounts(),
	}
}

func buildProviderDeployment(provider *pkgmetav1.Provider, revision v1.PackageRevision, cc *v1.ControllerConfig, namespace string, crdNames []string) (*corev1.ServiceAccount, *appsv1.Deployment) { // nolint:interfacer,gocyclo
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
					Containers:         getContainers(provider, revision, cc, namespace, crdNames),
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
*/
