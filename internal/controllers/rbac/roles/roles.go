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

package roles

import (
	"sort"

	rbacv1 "k8s.io/api/rbac/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/yndd/ndd-runtime/pkg/meta"
)

const (
	namePrefix       = "ndd:provider:"
	nameSuffixSystem = ":system"

	valTrue = "true"

	suffixStatus = "/status"

	pluralEvents            = "events"
	pluralConfigmaps        = "configmaps"
	pluralSecrets           = "secrets"
	pluralLeases            = "leases"
	pluralNetworkNodes      = "networknodes"
	pluralNetworkNodeUsages = "networknodeusages"
	pluralDeployments       = "deployments"
)

var (
	verbsEdit   = []string{rbacv1.VerbAll}
	verbsView   = []string{"get", "list", "watch"}
	verbsSystem = []string{"get", "list", "watch", "update", "patch", "create"}
)

// * Secrets for provider credentials and connection secrets.
// * ConfigMaps for leader election.
// * Leases for leader election.
// * Events for debugging.
// * NetworkNodes for make the ndd work
var rulesSystemExtra = []rbacv1.PolicyRule{
	{
		APIGroups: []string{"", "coordination/v1"},
		Resources: []string{pluralSecrets, pluralConfigmaps, pluralEvents, pluralLeases},
		Verbs:     verbsEdit,
	},
	{
		APIGroups: []string{"", "apiextensions.k8s.io"},
		Resources: []string{"*"},
		Verbs:     verbsEdit,
	},
	{
		APIGroups: []string{"", "dvr.ndd.yndd.io"},
		Resources: []string{pluralNetworkNodes},
		Verbs:     verbsView,
	},
	{
		APIGroups: []string{"", "dvr.ndd.yndd.io"},
		Resources: []string{pluralNetworkNodeUsages},
		Verbs:     verbsEdit,
	},
	{
		APIGroups: []string{"", "apps"},
		Resources: []string{pluralDeployments},
		Verbs:     verbsView,
	},
}

// SystemClusterRoleName returns the name of the 'system' cluster role - i.e.
// the role that a provider's ServiceAccount should be bound to.
func SystemClusterRoleName(revisionName string) string {
	return namePrefix + revisionName + nameSuffixSystem
}

// RenderClusterRoles returns ClusterRoles for the supplied ProviderRevision.
func RenderClusterRoles(pr *v1.ProviderRevision, crds []extv1.CustomResourceDefinition) []rbacv1.ClusterRole {
	// Our list of CRDs has no guaranteed order, so we sort them in order to
	// ensure we don't reorder our RBAC rules on each update.
	sort.Slice(crds, func(i, j int) bool { return crds[i].GetName() < crds[j].GetName() })

	groups := make([]string, 0)            // Allows deterministic iteration over groups.
	resources := make(map[string][]string) // Resources by group.
	for _, crd := range crds {
		if _, ok := resources[crd.Spec.Group]; !ok {
			resources[crd.Spec.Group] = make([]string, 0)
			groups = append(groups, crd.Spec.Group)
		}
		resources[crd.Spec.Group] = append(resources[crd.Spec.Group],
			crd.Spec.Names.Plural,
			crd.Spec.Names.Plural+suffixStatus,
		)
	}

	rules := []rbacv1.PolicyRule{}
	for _, g := range groups {
		rules = append(rules, rbacv1.PolicyRule{
			APIGroups: []string{g},
			Resources: resources[g],
		})
	}

	// The 'system' RBAC role does not aggregate; it is intended to be bound
	// directly to the service account that the provider runs as.
	system := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: SystemClusterRoleName(pr.GetName())},
		Rules:      append(append(withVerbs(rules, verbsSystem), rulesSystemExtra...), pr.Status.PermissionRequests...),
	}

	roles := []rbacv1.ClusterRole{*system}
	for i := range roles {
		ref := meta.AsController(meta.TypedReferenceTo(pr, v1.ProviderRevisionGroupVersionKind))
		roles[i].SetOwnerReferences([]metav1.OwnerReference{ref})
	}
	return roles
}

func withVerbs(r []rbacv1.PolicyRule, verbs []string) []rbacv1.PolicyRule {
	verbal := make([]rbacv1.PolicyRule, len(r))
	for i := range r {
		verbal[i] = r[i]
		verbal[i].Verbs = verbs
	}
	return verbal
}
