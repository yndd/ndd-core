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

package nn

import (
	"strings"

	ndddvrv1 "github.com/yndd/ndd-core/apis/dvr/v1"
	"github.com/yndd/ndd-runtime/pkg/meta"
	"github.com/yndd/ndd-runtime/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func buildDeployment(nn ndddvrv1.Nn, c *corev1.Container, namespace string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.Join([]string{ndddvrv1.PrefixDeployment, nn.GetName()}, "-"),
			Namespace: namespace,
			Labels: map[string]string{
				ndddvrv1.LabelNetworkDeviceDriver: strings.Join([]string{ndddvrv1.PrefixNetworkNode, nn.GetName()}, "-"),
			},
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(nn, ndddvrv1.NetworkNodeGroupVersionKind))},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: utils.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					ndddvrv1.LabelApplication: strings.Join([]string{ndddvrv1.PrefixNetworkNode, nn.GetName()}, "-"),
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						ndddvrv1.LabelApplication: strings.Join([]string{ndddvrv1.PrefixNetworkNode, nn.GetName()}, "-"),
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: strings.Join([]string{ndddvrv1.PrefixNetworkNode, nn.GetName()}, "-"),
					Containers: []corev1.Container{
						*c,
					},
				},
			},
		},
	}
}
