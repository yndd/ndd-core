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

package nn

import (
	"context"
	"strings"

	"github.com/yndd/ndd-runtime/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// Errors
	errEmptyTargetSecretReference   = "empty target secret reference"
	errCredentialSecretDoesNotExist = "credential secret does not exist"
	errEmptyTargetAddress           = "empty target address"
	errMissingUsername              = "missing username in credentials"
	errMissingPassword              = "missing password in credentials"
)

// Credentials holds the information for authenticating with the Server.
type Credentials struct {
	Username string
	Password string
}

func (v *NnValidator) ValidateCredentials(ctx context.Context, namespace, credentialsName, targetAddress string) (creds *Credentials, err error) {
	log := v.log.WithValues("namespace", namespace, "credentialsName", credentialsName, "targetAddress", targetAddress)
	log.Debug("Credentials Validation")
	// Retrieve the secret from Kubernetes for this network node
	if namespace == "" {
		namespace = "default"
	}
	credsSecret, err := v.GetSecret(ctx, namespace, credentialsName)
	if err != nil {
		return nil, err
	}

	// Check if address is defined on the network node
	if targetAddress == "" {
		return nil, errors.New(errEmptyTargetAddress)
	}

	creds = &Credentials{
		Username: strings.TrimSuffix(string(credsSecret.Data["username"]), "\n"),
		Password: strings.TrimSuffix(string(credsSecret.Data["password"]), "\n"),
	}

	log.Debug("Credentials", "creds", creds)

	if creds.Username == "" {
		return nil, errors.New(errMissingUsername)
	}
	if creds.Password == "" {
		return nil, errors.New(errMissingPassword)
	}

	return creds, nil
}

// Retrieve the secret containing the credentials for talking to the Network Node.
func (v *NnValidator) GetSecret(ctx context.Context, namespace, credentialsName string) (credsSecret *corev1.Secret, err error) {

	// check if credentialName is specified
	if credentialsName == "" {
		return nil, errors.New(errEmptyTargetSecretReference)
	}
	// check if credential secret exists
	secretKey := types.NamespacedName{
		Name:      credentialsName,
		Namespace: namespace,
	}
	credsSecret = &corev1.Secret{}
	if err := v.client.Get(ctx, secretKey, credsSecret); resource.IgnoreNotFound(err) != nil {
		return nil, errors.Wrap(err, errCredentialSecretDoesNotExist)
	}
	return credsSecret, nil
}
