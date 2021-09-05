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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	nddpkg "github.com/yndd/ndd-core/internal/nddpkg"
	"github.com/yndd/ndd-runtime/pkg/parser"
)

const (
	errGetNameFromMeta = "failed to get name from crossplane.yaml"
	errBuildPackage    = "failed to build package"
	errImageDigest     = "failed to get package digest"
	errCreatePackage   = "failed to create package file"
)

var packageRoot string
var ignore []string

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:          "build",
	Short:        "build a ndd package",
	Long:         "build a ndd package for usage with the network device driver in kubernetes",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := filepath.Abs(packageRoot)
		if err != nil {
			return err
		}
		// preprocess crd files
		// the further processing cannot handle --- in crd files
		var files []string
		if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			files = append(files, path)
			return nil
		}); err != nil {
			return err
		}
		for _, file := range files {
			input, err := ioutil.ReadFile(file)
			if err != nil {
				errors.Wrap(err, fmt.Sprintf("error reading file %s", file))
			}

			output := bytes.Replace(input, []byte("---"), []byte(""), -1)

			if err = ioutil.WriteFile(file, output, 0644); err != nil {
				errors.Wrap(err, fmt.Sprintf("error writing file %s", file))
			}
		}

		// process updates files
		metaScheme, err := nddpkg.BuildMetaScheme()
		if err != nil {
			return errors.New("cannot build meta scheme for package parser")
		}
		objScheme, err := nddpkg.BuildObjectScheme()
		if err != nil {
			return errors.New("cannot build object scheme for package parser")
		}
		img, err := nddpkg.Build(context.Background(),
			parser.NewFsBackend(buildChild.fs, parser.FsDir(root), parser.FsFilters(buildFilters(root, ignore)...)),
			parser.New(metaScheme, objScheme),
			buildChild.linter)
		if err != nil {
			return errors.Wrap(err, errBuildPackage)
		}

		hash, err := img.Digest()
		if err != nil {
			return errors.Wrap(err, errImageDigest)
		}
		pkgName := buildChild.name
		if pkgName == "" {
			metaPath := filepath.Join(root, nddpkg.MetaFile)
			pkgName, err = nddpkg.ParseNameFromMeta(buildChild.fs, metaPath)
			if err != nil {
				return errors.Wrap(err, errGetNameFromMeta)
			}
			pkgName = nddpkg.FriendlyID(pkgName, hash.Hex)
		}

		f, err := buildChild.fs.Create(nddpkg.BuildPath(root, pkgName))
		if err != nil {
			return errors.Wrap(err, errCreatePackage)
		}
		defer func() { _ = f.Close() }()
		return tarball.Write(nil, img, f)
	},
}

func init() {
	i := make([]string, 0)
	providerCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVarP(&packageRoot, "PackageRoot", "f", ".", "Path to package directory.")
	buildCmd.Flags().StringSliceVarP(&ignore, "Ignore", "", i, "Paths, specified relative to --package-root, to exclude from the package.")
	buildCmd.Flags().StringVarP(&packageName, "PackageName", "n", "", "Name of the package to be built. Uses name in ndd.yaml if not specified. Does not correspond to package tag.")

}

// default build filters skip directories, empty files, and files without YAML
// extension in addition to any paths specified.
func buildFilters(root string, skips []string) []parser.FilterFn {
	defaultFns := []parser.FilterFn{
		parser.SkipDirs(),
		parser.SkipNotYAML(),
		parser.SkipEmpty(),
	}
	opts := make([]parser.FilterFn, len(skips)+len(defaultFns))
	copy(opts, defaultFns)
	for i, s := range skips {
		opts[i+len(defaultFns)] = parser.SkipPath(filepath.Join(root, s))
	}
	return opts
}
