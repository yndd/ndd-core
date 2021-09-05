package clicmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	nddv1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:          "update",
	Short:        "update a ndd provider",
	Long:         "update a ndd provider for usage with the network device driver in kubernetes",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		k8sclopts := client.Options{
			Scheme: scheme,
		}
		c, err := client.New(config.GetConfigOrDie(), k8sclopts)
		if err != nil {
			return errors.Wrap(warnIfNotFound(err), errGetclient)
		}
		pKey := types.NamespacedName{
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

	},
}

func init() {
	providerCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVarP(&providerName, "providerName", "n", "", "Name of Provider.")
	updateCmd.Flags().StringVarP(&packageTag, "Tag", "t", "", "Tag of the package to be pushed. Must be a valid OCI image tag.")

}
