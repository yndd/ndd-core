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
	"os"
	"path/filepath"
	"strings"

	pkgmetav1 "github.com/yndd/ndd-core/apis/pkg/meta/v1"
	pkgv1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/yndd/ndd-runtime/pkg/meta"
	"github.com/yndd/ndd-runtime/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Options struct {
	serviceDiscoveryInfo  []*pkgv1.ServiceInfo
	grpcServiceName       string
	grpcCertSecretName    string
	compositeProviderName string
}

func renderProviderStatefulSet(pm *pkgmetav1.Provider, podSpec *pkgmetav1.PodSpec, pr pkgv1.PackageRevision, o *Options) *appsv1.StatefulSet {
	s := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            pr.GetName(),
			Namespace:       pm.Namespace,
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(pr, pkgv1.ProviderRevisionGroupVersionKind))},
		},
		Spec: appsv1.StatefulSetSpec{
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

func getEnv(o *Options) []corev1.EnvVar {
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
	envNodeName := corev1.EnvVar{
		Name: "NODE_NAME",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "spec.nodeName",
			},
		},
	}
	envNodeIP := corev1.EnvVar{
		Name: "NODE_IP",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.hostIP",
			},
		},
	}
	envGrpcSvc := corev1.EnvVar{
		Name:  "GRPC_SVC_NAME",
		Value: o.grpcServiceName,
	}

	certGrpcSecret := corev1.EnvVar{
		Name:  "GRPC_CERT_SECRET_NAME",
		Value: o.grpcCertSecretName,
	}

	svcDiscovery := corev1.EnvVar{
		Name:  "SERVICE_DISCOVERY",
		Value: os.Getenv("SERVICE_DISCOVERY"),
	}

	svcDiscoveryNamespace := corev1.EnvVar{
		Name:  "SERVICE_DISCOVERY_NAMESPACE",
		Value: os.Getenv("SERVICE_DISCOVERY_NAMESPACE"),
	}

	svcDiscoveryDCname := corev1.EnvVar{
		Name:  "SERVICE_DISCOVERY_DCNAME",
		Value: os.Getenv("SERVICE_DISCOVERY_DCNAME"),
	}

	envs := []corev1.EnvVar{
		envNameSpace,
		envPodIP,
		envPodName,
		envNodeName,
		envNodeIP,
		envGrpcSvc,
		certGrpcSecret,
		svcDiscovery,
		svcDiscoveryNamespace,
		svcDiscoveryDCname,
	}

	for _, serviceInfo := range o.serviceDiscoveryInfo {
		switch serviceInfo.Kind {
		case pkgv1.KindNone:
			envs = append(envs, corev1.EnvVar{
				Name:  "TARGET_SERVICE_NAME",
				Value: serviceInfo.ServiceName,
			})
		default:
			envs = append(envs, corev1.EnvVar{
				Name:  "SERVICE_NAME",
				Value: serviceInfo.ServiceName,
			})
		}
	}

	if o.compositeProviderName != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "COMPOSITE_PROVIDER_NAME",
			Value: o.compositeProviderName,
		})
	}

	return envs
}

func getContainers(p *pkgmetav1.Provider, podSpec *pkgmetav1.PodSpec, pullPolicy *corev1.PullPolicy, o *Options) []corev1.Container {
	containers := []corev1.Container{}

	for _, c := range podSpec.Containers {
		if c.Container.Name == "kube-rbac-proxy" {
			containers = append(containers, getKubeProxyContainer(c))
		} else {
			containers = append(containers, getContainer(p, c, pullPolicy, o))
		}
	}

	return containers
}

func getKubeProxyContainer(c *pkgmetav1.ContainerSpec) corev1.Container {
	return corev1.Container{
		Name:  c.Container.Name,
		Image: c.Container.Image,
		Args:  getProxyArgs(),
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: 8443,
				Name:          "https",
			},
		},
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

func getArgs(pm *pkgmetav1.Provider) []string {
	args := []string{"start"}
	switch pm.Spec.Pod.Type {
	case pkgmetav1.DeploymentTypeStatefulset:
	case pkgmetav1.DeploymentTypeDeployment:
	default:
	}

	args = append(args, "--debug")
	return args
}

func getVolumeMounts(c *pkgmetav1.ContainerSpec) []corev1.VolumeMount {
	volumes := []corev1.VolumeMount{}
	for _, extra := range c.Extras {
		if extra.Certificate {
			volumes = append(volumes, corev1.VolumeMount{
				Name:      strings.Join([]string{c.Container.Name, extra.Name}, "-"),
				MountPath: filepath.Join("tmp", strings.Join([]string{"k8s", extra.Name, "server"}, "-"), certPathSuffix),
				ReadOnly:  true,
			})
		} else {
			if extra.Volume {
				volumes = append(volumes, corev1.VolumeMount{
					Name:      strings.Join([]string{c.Container.Name, extra.Name}, "-"),
					MountPath: filepath.Join(extra.Name),
				})
			}
		}
	}
	return volumes
}

func getVolumes(podSpec *pkgmetav1.PodSpec, pr pkgv1.PackageRevision) []corev1.Volume {
	volume := []corev1.Volume{}
	for _, c := range podSpec.Containers {
		for _, extra := range c.Extras {
			if extra.Certificate {
				volume = append(volume, corev1.Volume{
					Name: strings.Join([]string{c.Container.Name, extra.Name}, "-"),
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  getCertificateName(pr.GetName(), c.Container.Name, extra.Name),
							DefaultMode: utils.Int32Ptr(420),
						},
					},
				})
			} else {
				volume = append(volume, corev1.Volume{
					Name: strings.Join([]string{c.Container.Name, extra.Name}, "-"),
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				})
			}
		}
	}
	return volume
}

func getContainer(p *pkgmetav1.Provider, c *pkgmetav1.ContainerSpec, pullPolicy *corev1.PullPolicy, o *Options) corev1.Container {
	return corev1.Container{
		Name:            c.Container.Name,
		Image:           c.Container.Image,
		ImagePullPolicy: *pullPolicy,
		SecurityContext: getSecurityContext(),
		Args:            getArgs(p),
		Env:             getEnv(o),
		Command: []string{
			containerStartupCmd,
		},
		VolumeMounts: getVolumeMounts(c),
	}
}
