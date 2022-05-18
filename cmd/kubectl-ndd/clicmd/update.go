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

package clicmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	nddv1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:          "update",
	Short:        "update a ndd package",
	Long:         "update a ndd package for usage with the network device driver in kubernetes",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		k8sclopts := client.Options{
			Scheme: scheme,
		}
		c, err := client.New(config.GetConfigOrDie(), k8sclopts)
		if err != nil {
			return errors.Wrap(warnIfNotFound(err), errGetclient)
		}
		var pKey types.NamespacedName
		if providerName != "" {
			pKey = types.NamespacedName{
				Namespace: "default",
				Name:      providerName,
			}
			preProv := &nddv1.Provider{}
			if err := c.Get(context.Background(), pKey, preProv); err != nil {
				return errors.Wrap(warnIfNotFound(err), "cannot update provider")
			}
			pkg := preProv.Spec.Package
			pkgReference, err := name.ParseReference(pkg, name.WithDefaultRegistry(""))
			if err != nil {
				return errors.Wrap(warnIfNotFound(err), "cannot update provider")
			}
			newPkg := ""
			if strings.HasPrefix(packageTag, "sha256") {
				newPkg = pkgReference.Context().Digest(packageTag).Name()
			} else {
				newPkg = pkgReference.Context().Tag(packageTag).Name()
			}
			preProv.Spec.Package = newPkg
			//req, err := json.Marshal(preProv)
			//if err != nil {
			//	return errors.Wrap(warnIfNotFound(err), "cannot update provider")
			//}
			if err := c.Update(context.Background(), preProv); err != nil {
				return errors.Wrap(warnIfNotFound(err), "cannot update provider")
			}

			//res, err := kube.Providers().Patch(context.Background(), providerName, types.MergePatchType, req, metav1.PatchOptions{})
			//if err != nil {
			//}
			_, err = fmt.Fprintf(os.Stdout, "%s/%s updated\n", strings.ToLower(nddv1.ProviderGroupKind), providerName)
			return err
		}
		if intentName != "" {
			pKey = types.NamespacedName{
				Namespace: "default",
				Name:      intentName,
			}
			preInt := &nddv1.Intent{}
			if err := c.Get(context.Background(), pKey, preInt); err != nil {
				return errors.Wrap(warnIfNotFound(err), "cannot update intent")
			}
			pkg := preInt.Spec.Package
			pkgReference, err := name.ParseReference(pkg, name.WithDefaultRegistry(""))
			if err != nil {
				return errors.Wrap(warnIfNotFound(err), "cannot update intent")
			}
			newPkg := ""
			if strings.HasPrefix(packageTag, "sha256") {
				newPkg = pkgReference.Context().Digest(packageTag).Name()
			} else {
				newPkg = pkgReference.Context().Tag(packageTag).Name()
			}
			preInt.Spec.Package = newPkg
			//req, err := json.Marshal(preProv)
			//if err != nil {
			//	return errors.Wrap(warnIfNotFound(err), "cannot update provider")
			//}
			if err := c.Update(context.Background(), preInt); err != nil {
				return errors.Wrap(warnIfNotFound(err), "cannot update intent")
			}

			//res, err := kube.Providers().Patch(context.Background(), providerName, types.MergePatchType, req, metav1.PatchOptions{})
			//if err != nil {
			//}
			_, err = fmt.Fprintf(os.Stdout, "%s/%s updated\n", strings.ToLower(nddv1.ProviderGroupKind), intentName)
			return err
		}
		return nil
	},
}

func init() {
	packageCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVarP(&providerName, "providerName", "", "", "Name of Provider.")
	updateCmd.Flags().StringVarP(&intentName, "intentName", "", "", "Name of Intent.")
	updateCmd.Flags().StringVarP(&packageTag, "Tag", "t", "", "Tag of the package to be pushed. Must be a valid OCI image tag.")

}
