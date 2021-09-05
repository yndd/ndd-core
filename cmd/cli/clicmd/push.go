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
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	nddpkg "github.com/yndd/ndd-core/internal/nddpkg"
)

const (
	errGetwd           = "failed to get working directory while searching for package"
	errFindPackageinWd = "failed to find a package current working directory"
)

var (
	nddPackageName string
	packageTag     string
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:          "push",
	Short:        "push a ndd provider",
	Long:         "push a ndd provider for usage with the network device driver in kubernetes",
	SilenceUsage: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a Tag of the package to be pushed. Must be a valid OCI image tag.")
		}
		packageTag = args[0]
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		tag, err := name.NewTag(packageTag)
		if err != nil {
			return err
		}

		// If package is not defined, attempt to find single package in current
		// directory.
		if nddPackageName == "" {
			wd, err := os.Getwd()
			if err != nil {
				return errors.Wrap(err, errGetwd)
			}
			path, err := nddpkg.FindNddpkgInDir(pushChild.fs, wd)
			if err != nil {
				return errors.Wrap(err, errFindPackageinWd)
			}
			nddPackageName = path
		}
		img, err := tarball.ImageFromPath(nddPackageName, nil)
		if err != nil {
			return err
		}
		return remote.Write(tag, img, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	},
}

func init() {
	providerCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringVarP(&nddPackageName, "NddPackageName", "p", "", "Path to package. If not specified and only one package exists in current directory it will be used.")
}
