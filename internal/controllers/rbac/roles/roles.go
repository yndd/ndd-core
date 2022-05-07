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

package roles

import (
	"fmt"
	"sort"

	rbacv1 "k8s.io/api/rbac/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/yndd/ndd-runtime/pkg/meta"
)

const (
	nameProviderPrefix       = "ndd:provider:"
	nameIntentPrefix         = "nddo:intent:"
	nameProviderMetricPrefix = "ndd:provider:metrics:"
	nameIntentMetricPrefix   = "nddo:intent:metrics:"
	nameSuffixSystem         = ":system"

	//valTrue = "true"

	suffixStatus = "/status"

	pluralEvents            = "events"
	pluralConfigmaps        = "configmaps"
	pluralSecrets           = "secrets"
	pluralLeases            = "leases"
	pluralServices          = "services"
	pluralServiceAccounts   = "serviceaccounts"
	pluralNetworkNodes      = "networknodes"
	pluralNetworkNodeUsages = "networknodeusages"
	pluralDeployments       = "deployments"
	pluralStatefulsets      = "statefulsets"
	pluralPods              = "pods"
	pluralCrds              = "customresourcedefinitions"
)

var (
	verbsEdit   = []string{rbacv1.VerbAll}
	verbsView   = []string{"get", "list", "watch"}
	verbsSystem = []string{"get", "list", "watch", "update", "patch", "create", "delete"}
)

var rulesSystemExtraNew = []rbacv1.PolicyRule{
	{
		APIGroups: []string{"*"},
		Resources: []string{pluralPods},
		Verbs:     verbsView,
	},
	{
		APIGroups: []string{"*"},
		Resources: []string{pluralServices, pluralServiceAccounts},
		Verbs:     verbsEdit,
	},
	{

		APIGroups: []string{"rbac.authorization.k8s.io"},
		Resources: []string{"clusterroles", "clusterrolebindings"},
		Verbs:     verbsEdit,
	},
	{
		APIGroups: []string{"coordination/v1"},
		Resources: []string{pluralSecrets, pluralConfigmaps, pluralEvents, pluralLeases},
		Verbs:     verbsEdit,
	},
	{
		APIGroups: []string{"apps"},
		Resources: []string{pluralDeployments, pluralStatefulsets},
		Verbs:     verbsEdit,
	},
	{
		APIGroups: []string{"apiextensions.k8s.io"},
		Resources: []string{pluralCrds},
		Verbs:     verbsView,
	},
	{
		APIGroups: []string{"meta.pkg.ndd.yndd.io"},
		Resources: []string{"providers"},
		Verbs:     verbsEdit,
	},
	{
		APIGroups: []string{"pkg.ndd.yndd.io"},
		Resources: []string{"providerrevisions", "providers"},
		Verbs:     verbsView,
	},
	{
		APIGroups: []string{"cert-manager.io"},
		Resources: []string{"certificates"},
		Verbs:     verbsEdit,
	},
	{
		APIGroups: []string{"admissionregistration.k8s.io"},
		Resources: []string{"mutatingwebhookconfigurations", "validatingwebhookconfigurations"},
		Verbs:     verbsEdit,
	},
}

// * Secrets for provider credentials and connection secrets.
// * ConfigMaps for leader election.
// * Leases for leader election.
// * Events for debugging.
// * NetworkNodes for make the ndd work
var rulesSystemExtra = []rbacv1.PolicyRule{
	{
		APIGroups: []string{"*"},
		Resources: []string{pluralSecrets, pluralConfigmaps, pluralEvents, pluralLeases},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"coordination/v1"},
		Resources: []string{pluralSecrets, pluralConfigmaps, pluralEvents, pluralLeases},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"apiextensions.k8s.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"dvr.ndd.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"meta.pkg.ndd.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"apps"},
		Resources: []string{pluralDeployments},
		Verbs:     verbsView,
	},
	{
		APIGroups: []string{"nddo.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"topo.nddr.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"ipam.nddr.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"aspool.nddr.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"vlanpool.nddr.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"nipool.nddr.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"esipool.nddr.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"org.nddr.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"network.ndda.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"srl.nddp.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"srl3.nddp.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"sros.nddp.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"srl.ndda.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"sros.ndda.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"as.nddr.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"ni.nddr.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"esi.nddr.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"rt.nddr.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
	{
		APIGroups: []string{"vlan.nddr.yndd.io"},
		Resources: []string{"*"},
		Verbs:     verbsSystem,
	},
}

// SystemClusterProviderRoleName returns the name of the 'system' cluster role - i.e.
// the role that a provider's ServiceAccount should be bound to.
func SystemClusterProviderRoleName(revisionName string) string {
	return nameProviderPrefix + revisionName + nameSuffixSystem
}

// SystemClusterIntentRoleName returns the name of the 'system' cluster role - i.e.
// the role that a intent's ServiceAccount should be bound to.
func SystemClusterIntentRoleName(revisionName string) string {
	return nameIntentPrefix + revisionName + nameSuffixSystem
}

// SystemClusterProviderMetricRoleName returns the name of the 'system' cluster role - i.e.
// the role that a provider's ServiceAccount should be bound to.
func SystemClusterProviderMetricRoleName(revisionName string) string {
	return nameProviderMetricPrefix + revisionName + nameSuffixSystem
}

// SystemClusterIntentMetricRoleName returns the name of the 'system' cluster role - i.e.
// the role that a intent's ServiceAccount should be bound to.
func SystemClusterIntentMetricRoleName(revisionName string) string {
	return nameIntentMetricPrefix + revisionName + nameSuffixSystem
}

// RenderClusterRoles returns ClusterRoles for the supplied PackageRevision.
func RenderClusterRoles(pr *v1.PackageRevision, crds []extv1.CustomResourceDefinition) []rbacv1.ClusterRole {
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

	//fmt.Printf("rules: %v\n", rules)

	// The 'system' RBAC role does not aggregate; it is intended to be bound
	// directly to the service account that the provider/intent runs as.
	var system *rbacv1.ClusterRole
	switch (*pr).GetKind() {
	case v1.IntentRevisionKind:
		// Intent revision
		system = &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{Name: SystemClusterIntentRoleName((*pr).GetRevName())},
			Rules:      append(append(withVerbs(rules, verbsSystem), rulesSystemExtra...), (*pr).GetPermissionsRequests()...),
		}
	default:
		// Provider revision
		/*
			pmrs := []rbacv1.PolicyRule{}
			for _, pmr := range (*pr).GetPermissionsRequests() {
				if len(resources) == 0 {
					pmr.Resources = []string{"*"}
				}
				pmrs = append(pmrs, pmr)
			}
		*/

		fmt.Printf("pmrs: %v\n", (*pr).GetPermissionsRequests())
		system = &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{Name: SystemClusterProviderRoleName((*pr).GetRevName())},
			Rules:      append(append(withVerbs(rules, verbsSystem), rulesSystemExtraNew...), (*pr).GetPermissionsRequests()...),
		}
	}

	roles := []rbacv1.ClusterRole{*system}
	for i := range roles {
		var ref metav1.OwnerReference

		switch (*pr).GetKind() {
		case v1.IntentRevisionKind:
			// Intent revision
			ref = meta.AsController(meta.TypedReferenceTo(*pr, v1.IntentRevisionGroupVersionKind))
		default:
			// Provider revision
			ref = meta.AsController(meta.TypedReferenceTo(*pr, v1.ProviderRevisionGroupVersionKind))
		}
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
