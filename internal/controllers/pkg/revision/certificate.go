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
	certv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pkgmetav1 "github.com/yndd/ndd-core/apis/pkg/meta/v1"
	pkgv1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/yndd/ndd-runtime/pkg/meta"
)

func renderCertificate(p *pkgmetav1.Provider, podSpec *pkgmetav1.PodSpec, c *pkgmetav1.ContainerSpec, extra *pkgmetav1.Extras, pr pkgv1.PackageRevision) *certv1.Certificate { // nolint:interfacer,gocyclo
	certificateName := getCertificateName(pr.GetName(), c.Container.Name, extra.Name)
	serviceName := getServiceName(pr.GetLabels()[pkgv1.ParentLabelKey], c.Container.Name, extra.Name)
	servicePrName := getServiceName(pr.GetName(), c.Container.Name, extra.Name)

	//newCertServiceName := getServiceName(strings.Join([]string{pr.GetName(), "0"}, "-"), c.Container.Name, extra.Name)

	return &certv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      certificateName,
			Namespace: p.Namespace,
			Labels: map[string]string{
				getLabelKey(extra.Name): serviceName,
			},
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(pr, pkgv1.ProviderRevisionGroupVersionKind))},
		},
		Spec: certv1.CertificateSpec{
			DNSNames: []string{
				//strings.Join([]string{pr.GetName(), "0"}, "-"),
				//strings.Join([]string{pr.GetName(), "0", "cluster", "local"}, "-"),
				//getDnsName(p.Namespace, pr.GetName()+"-0"),
				//getDnsName(p.Namespace, pr.GetName()+"-0", "cluster", "local"),
				getDnsName(p.Namespace, serviceName),
				getDnsName(p.Namespace, serviceName, "cluster", "local"),
				getDnsName(p.Namespace, servicePrName),
				getDnsName(p.Namespace, servicePrName, "cluster", "local"),
				//getDnsName(p.Namespace, newCertServiceName),
				//getDnsName(p.Namespace, newCertServiceName, "cluster", "local"),
			},
			IssuerRef: certmetav1.ObjectReference{
				Kind: "Issuer",
				Name: "selfsigned-issuer",
			},
			SecretName: certificateName,
		},
	}
}
