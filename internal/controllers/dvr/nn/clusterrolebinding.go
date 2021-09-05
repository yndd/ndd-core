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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// items
	kindClusterRole = "ClusterRole"
	clusterRoleName = "ndd-manager-role"
)

func buildClusterRoleBinding(nn ndddvrv1.Nn, namespace string) *rbacv1.ClusterRoleBinding {
	subjects := make([]rbacv1.Subject, 0)
	subjects = append(subjects, rbacv1.Subject{
		Kind:      rbacv1.ServiceAccountKind,
		Namespace: namespace,
		Name:      strings.Join([]string{ndddvrv1.PrefixNetworkNode, nn.GetName()}, "-"),
	})

	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.Join([]string{ndddvrv1.PrefixNetworkNode, nn.GetName()}, "-"),
			Namespace: namespace,
			Labels: map[string]string{
				ndddvrv1.LabelNetworkDeviceDriver: strings.Join([]string{ndddvrv1.PrefixNetworkNode, nn.GetName()}, "-"),
			},
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(nn, ndddvrv1.NetworkNodeGroupVersionKind))},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     kindClusterRole,
			Name:     clusterRoleName,
		},
		Subjects: subjects,
	}

}
