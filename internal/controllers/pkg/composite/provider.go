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

package composite

import (
	"strings"

	pkgv1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/yndd/ndd-runtime/pkg/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getProviderName(compositeProviderName, pkgName string) string {
	return strings.Join([]string{compositeProviderName, pkgName}, "-")
}

func renderProvider(cp *pkgv1.CompositeProvider, pkg pkgv1.PackageSpec) *pkgv1.Provider {
	return &pkgv1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getProviderName(cp.Name, pkg.Name),
			Namespace: cp.Namespace,
			Labels: map[string]string{
				strings.Join([]string{pkgv1.Group, "composite-provider-name"}, "/"):      cp.Name,
				strings.Join([]string{pkgv1.Group, "composite-provider-namespace"}, "/"): cp.Namespace,
			},
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(cp, pkgv1.CompositeProviderGroupVersionKind))},
		},
		Spec: pkgv1.ProviderSpec{
			PackageSpec: pkg,
		},
	}

}
