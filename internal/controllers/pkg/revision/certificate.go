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
	"strings"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	certmetav1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pkgmetav1 "github.com/yndd/ndd-core/apis/pkg/meta/v1"
	v1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/yndd/ndd-runtime/pkg/meta"
)

func buildProviderWebhookCertificate(provider *pkgmetav1.Provider, revision v1.PackageRevision, namespace string) *certv1.Certificate { // nolint:interfacer,gocyclo
	webhookCertificateName := strings.Join([]string{revision.GetName(), "webhook", "serving-cert"}, "-")
	webhookServiceName := strings.Join([]string{revision.GetName(), "webhook", "svc"}, "-")
	return &certv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      webhookCertificateName,
			Namespace: namespace,
			Labels: map[string]string{
				"webhook": webhookCertificateName,
			},
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(revision, v1.ProviderRevisionGroupVersionKind))},
		},
		Spec: certv1.CertificateSpec{
			DNSNames: []string{
				strings.Join([]string{webhookServiceName, namespace, "svc"}, "."),
				strings.Join([]string{webhookServiceName, namespace, "svc", "cluster", "local"}, "."),
			},
			IssuerRef: certmetav1.ObjectReference{
				Kind: "Issuer",
				Name: "selfsigned-issuer",
			},
			SecretName: webhookCertificateName,
		},
	}
}

func buildProviderGnmiCertificate(provider *pkgmetav1.Provider, revision v1.PackageRevision, namespace string) *certv1.Certificate { // nolint:interfacer,gocyclo
	gnmiCertificateName := strings.Join([]string{revision.GetName(), "gnmi", "serving-cert"}, "-")
	gnmiServiceName := strings.Join([]string{revision.GetName(), "gnmi", "svc"}, "-")
	return &certv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gnmiCertificateName,
			Namespace: namespace,
			Labels: map[string]string{
				"gnmi": gnmiCertificateName,
			},
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(revision, v1.ProviderRevisionGroupVersionKind))},
		},
		Spec: certv1.CertificateSpec{
			DNSNames: []string{
				strings.Join([]string{gnmiServiceName, namespace, "svc"}, "."),
				strings.Join([]string{gnmiServiceName, namespace, "svc", "cluster", "local"}, "."),
			},
			IssuerRef: certmetav1.ObjectReference{
				Kind: "Issuer",
				Name: "selfsigned-issuer",
			},
			SecretName: gnmiCertificateName,
		},
	}
}
