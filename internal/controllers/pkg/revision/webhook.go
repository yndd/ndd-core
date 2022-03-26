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
	"strings"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pkgmetav1 "github.com/yndd/ndd-core/apis/pkg/meta/v1"
	v1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/yndd/ndd-runtime/pkg/meta"
	"github.com/yndd/ndd-runtime/pkg/utils"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func buildProviderWebhookMutate(provider *pkgmetav1.Provider, revision v1.PackageRevision, namespace string, crd *extv1.CustomResourceDefinition) *admissionv1.MutatingWebhookConfiguration { // nolint:interfacer,gocyclo
	webhookCertificateName := strings.Join([]string{revision.GetName(), "webhook", "serving-cert"}, "-")
	webhookServiceName := strings.Join([]string{revision.GetName(), "webhook", "svc"}, "-")

	//o := strings.Split(provider.GetName(), "-")

	v := getVersions(crd.Spec.Versions)

	fmt.Printf("webhook versions: %v\n", v)

	//fmt.Printf("crd group: %s, versions: %v, singularName: %s, pluralName: %s \n", crd.Spec.Group, versions, crd.Spec.Names.Singular, crd.Spec.Names.Plural)
	//+kubebuilder:webhook:path=/mutate-srl3-nddp-yndd-io,mutating=true,failurePolicy=fail,sideEffects=None,groups=srl3.nddp.yndd.io,resources="*",verbs=create;update,versions=v1alpha1,name=mutate.srl3.nddp.yndd.io,admissionReviewVersions=v1

	failurePolicy := admissionv1.Fail
	sideEffect := admissionv1.SideEffectClassNone
	return &admissionv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.Join([]string{provider.GetName(), "mutating-webhook-configuration"}, "-"),
			Annotations: map[string]string{
				"cert-manager.io/inject-ca-from": strings.Join([]string{namespace, webhookCertificateName}, "/"),
			},
			//Namespace:       namespace,
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(revision, v1.ProviderRevisionGroupVersionKind))},
		},
		Webhooks: []admissionv1.MutatingWebhook{
			{
				//Name:                    "msrl3device.srl3.nddp.yndd.io",
				//Name:                    strings.Join([]string{"mutate", crd.Spec.Group}, "."),
				Name:                    strings.Join([]string{"m" + crd.Spec.Names.Singular, crd.Spec.Group}, "."),
				AdmissionReviewVersions: []string{"v1"},
				ClientConfig: admissionv1.WebhookClientConfig{
					Service: &admissionv1.ServiceReference{
						Name:      webhookServiceName,
						Namespace: namespace,
						// orig
						//Path:      utils.StringPtr("/mutate-srl3-nddp-yndd-io-v1alpha1-srl3device"),
						// new
						//Path:      utils.StringPtr("/mutate-srl3-nddp-yndd-io"),
						Path: utils.StringPtr(strings.Join([]string{"/mutate", strings.ReplaceAll(crd.Spec.Group, ".", "-"), v[0], crd.Spec.Names.Singular}, "-")),
					},
				},
				Rules: []admissionv1.RuleWithOperations{
					{
						Rule: admissionv1.Rule{
							//APIGroups:   []string{strings.Join([]string{o[len(o)-1], "nddp.yndd.io"}, ".")},
							APIGroups:   []string{crd.Spec.Group},
							APIVersions: v,
							Resources:   []string{crd.Spec.Names.Plural},
						},
						Operations: []admissionv1.OperationType{
							admissionv1.Create,
							admissionv1.Update,
						},
					},
				},
				FailurePolicy: &failurePolicy,
				SideEffects:   &sideEffect,
			},
		},
	}
}

func buildProviderWebhookValidate(provider *pkgmetav1.Provider, revision v1.PackageRevision, namespace string, crd *extv1.CustomResourceDefinition) *admissionv1.ValidatingWebhookConfiguration { // nolint:interfacer,gocyclo
	webhookCertificateName := strings.Join([]string{revision.GetName(), "webhook", "serving-cert"}, "-")
	webhookServiceName := strings.Join([]string{revision.GetName(), "webhook", "svc"}, "-")

	//o := strings.Split(provider.GetName(), "-")

	v := getVersions(crd.Spec.Versions)
	fmt.Printf("webhook versions: %v\n", v)

	failurePolicy := admissionv1.Fail
	sideEffect := admissionv1.SideEffectClassNone
	return &admissionv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.Join([]string{provider.GetName(), "validating-webhook-configuration"}, "-"),
			Annotations: map[string]string{
				"cert-manager.io/inject-ca-from": strings.Join([]string{namespace, webhookCertificateName}, "/"),
			},
			//Namespace:       namespace,
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(revision, v1.ProviderRevisionGroupVersionKind))},
		},
		Webhooks: []admissionv1.ValidatingWebhook{
			{
				//srl3devices.srl3.nddp.yndd.io
				//Name:                    "vsrl3device.srl3.nddp.yndd.io",
				Name:                    strings.Join([]string{"v" + crd.Spec.Names.Singular, crd.Spec.Group}, "."),
				AdmissionReviewVersions: []string{"v1"},
				ClientConfig: admissionv1.WebhookClientConfig{
					Service: &admissionv1.ServiceReference{
						Name:      webhookServiceName,
						Namespace: namespace,
						//Path:      utils.StringPtr("/validate-srl3-nddp-yndd-io-v1alpha1-srl3device"),
						Path: utils.StringPtr(strings.Join([]string{"/validate", strings.ReplaceAll(crd.Spec.Group, ".", "-"), v[0], crd.Spec.Names.Singular}, "-")),
					},
				},
				Rules: []admissionv1.RuleWithOperations{
					{
						Rule: admissionv1.Rule{
							//APIGroups:   []string{strings.Join([]string{o[len(o)-1], "nddp.yndd.io"}, ".")},
							//APIVersions: []string{"v1alpha1"},
							//Resources:   []string{"*"},
							APIGroups:   []string{crd.Spec.Group},
							APIVersions: v,
							Resources:   []string{crd.Spec.Names.Plural},
						},
						Operations: []admissionv1.OperationType{
							admissionv1.Create,
							admissionv1.Update,
						},
					},
				},
				FailurePolicy: &failurePolicy,
				SideEffects:   &sideEffect,
			},
		},
	}
}

func getVersions(crdVersions []extv1.CustomResourceDefinitionVersion) []string {
	versions := []string{}
	for _, crdVersion := range crdVersions {
		fmt.Printf("crdversion name: %v\n", crdVersion.Name)
		found := false
		for _, version := range versions {
			if crdVersion.Name == version {
				found = true
				break
			}

		}
		if !found {
			versions = append(versions, crdVersion.Name)
		}
	}
	return versions
}
