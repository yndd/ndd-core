package clicmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	nddv1 "github.com/yndd/ndd-core/apis/pkg/v1"
	nddpkg "github.com/yndd/ndd-core/internal/nddpkg"
	"github.com/yndd/ndd-core/internal/version"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	errPkgIdentifier = "invalid package image identifier"
	errGetclient     = "cannot get k8s client"
)

//var packageName string
//var providerName string
var revisionHistoryLimit int64
var PackagePullSecrets []string
var manualActivation bool

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:          "install",
	Short:        "install a ndd package",
	Long:         "install a ndd package for usage with the network device driver in kubernetes",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		rap := nddv1.AutomaticActivation
		if manualActivation {
			rap = nddv1.ManualActivation
		}
		pkgName := providerName
		if pkgName == "" {
			ref, err := name.ParseReference(packageName)
			if err != nil {
				return errors.Wrap(err, errPkgIdentifier)
			}
			pkgName = nddpkg.ToDNSLabel(ref.Context().RepositoryStr())
		}
		packagePullSecrets := make([]corev1.LocalObjectReference, len(PackagePullSecrets))
		for i, s := range PackagePullSecrets {
			packagePullSecrets[i] = corev1.LocalObjectReference{
				Name: s,
			}
		}
		cr := &nddv1.Provider{
			ObjectMeta: metav1.ObjectMeta{
				Name: pkgName,
			},
			Spec: nddv1.ProviderSpec{
				PackageSpec: nddv1.PackageSpec{
					Package:                  packageName,
					RevisionActivationPolicy: &rap,
					RevisionHistoryLimit:     &revisionHistoryLimit,
					PackagePullSecrets:       packagePullSecrets,
				},
			},
		}
		fmt.Printf("cr %v", cr)
		k8sclopts := client.Options{
			Scheme: scheme,
		}
		c, err := client.New(config.GetConfigOrDie(), k8sclopts)
		if err != nil {
			return errors.Wrap(warnIfNotFound(err), errGetclient)
		}

		if err := c.Create(context.Background(), cr); err != nil {
			return errors.Wrap(warnIfNotFound(err), "cannot create provider")
		}

		_, err = fmt.Fprintf(os.Stdout, "%s/%s created\n", strings.ToLower(nddv1.ProviderGroupKind), pkgName)
		return err
	},
}

func init() {
	i := make([]string, 0)
	providerCmd.AddCommand(installCmd)
	installCmd.PersistentFlags().StringVarP(&packageName, "PackageName", "p", "", "Image containing Provider package.")
	installCmd.Flags().StringVarP(&providerName, "providerName", "n", "", "Name of Provider.")
	installCmd.Flags().Int64VarP(&revisionHistoryLimit, "RevisionHistoryLimit", "r", 1, "Revision history limit.")
	installCmd.Flags().BoolVarP(&manualActivation, "ManualActivation", "", false, "Enable manual revision activation policy")
	installCmd.Flags().StringSliceVarP(&PackagePullSecrets, "PackagePullSecrets", "", i, "List of secrets used to pull package.")
}

func warnIfNotFound(err error) error {
	serr, ok := err.(*apierrors.StatusError)
	if !ok {
		return err
	}
	if serr.ErrStatus.Code != http.StatusNotFound {
		return err
	}
	return errors.WithMessagef(err, "kubectl-ndd plugin %s might be out of date", version.New().GetVersionString())
}
