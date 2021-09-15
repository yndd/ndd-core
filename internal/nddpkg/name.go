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

package nddpkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"

	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"
)

const (
	// MetaFile is the name of a Ndd package metadata file.
	MetaFile string = "ndd.yaml"

	// StreamFile is the name of the file in a ndd package image that
	// contains its YAML stream.
	StreamFile string = "package.yaml"

	// StreamFileMode determines the permissions on the stream file.
	StreamFileMode os.FileMode = 0o644

	// NddpkgExtension is the extension for compiled ndd packages.
	NddpkgExtension string = ".nddpkg"

	// NddpkgMatchPattern is the match pattern for identifying compiled ndd packages.
	NddpkgMatchPattern string = "*" + NddpkgExtension

	// identifierDelimeters is the set of valid OCI image identifier delimeter
	// characters.
	identifierDelimeters string = ":@"
)

func truncate(str string, num int) string {
	t := str
	if len(str) > num {
		t = str[0:num]
	}
	return t
}

// FriendlyID builds a valid DNS label string made up of the name of a package
// and its image digest.
func FriendlyID(name, hash string) string {
	return ToDNSLabel(strings.Join([]string{truncate(name, 50), truncate(hash, 12)}, "-"))
}

// ToDNSLabel converts the string to a valid DNS label.
func ToDNSLabel(s string) string { // nolint:gocyclo
	var cut strings.Builder
	for i := range s {
		b := s[i]
		if ('a' <= b && b <= 'z') || ('0' <= b && b <= '9') {
			cut.WriteByte(b)
		}
		if (b == '.' || b == '/' || b == ':' || b == '-') && (i != 0 && i != 62 && i != len(s)-1) {
			cut.WriteByte('-')
		}
		if i == 62 {
			break
		}
	}
	return strings.Trim(cut.String(), "-")
}

// BuildPath builds a path for a compiled Ndd package. If file name has
// extension it will be replaced.
func BuildPath(path, name string) string {
	full := filepath.Join(path, name)
	ext := filepath.Ext(full)
	return full[0:len(full)-len(ext)] + NddpkgExtension
}

// ParseNameFromMeta extracts the package name from its meta file.
func ParseNameFromMeta(fs afero.Fs, path string) (string, error) {
	bs, err := afero.ReadFile(fs, filepath.Clean(path))
	if err != nil {
		return "", err
	}
	pkgName, err := parseNameFromPackage(bs)
	if err != nil {
		return "", err
	}
	return pkgName, nil
}

// ParseKindFromMeta extracts the package kind from its meta file.
func ParseKindFromMeta(fs afero.Fs, path string) (string, error) {
	bs, err := afero.ReadFile(fs, filepath.Clean(path))
	if err != nil {
		return "", err
	}
	pkgKind, err := parseKindFromPackage(bs)
	if err != nil {
		return "", err
	}
	return pkgKind, nil
}

// ParsePackageSourceFromReference parses a package source from an OCI image
// reference. A source is defined as an OCI image reference with the identifier
// (tag or digest) stripped and no other changes to the original reference
// source. This is necessary because go-containerregistry will convert docker.io
// to index.docker.io for backwards compatibility before pulling an image. We do
// not want to do that in cases where we are not pulling an image because it
// breaks comparison with dependencies defined in a Configuration manifest.
func ParsePackageSourceFromReference(ref name.Reference) string {
	return strings.TrimRight(strings.TrimSuffix(ref.String(), ref.Identifier()), identifierDelimeters)
}

type metaPkg struct {
	Metadata struct {
		Name string `json:"name"`
		Kind string `json:"kind"`
	}
}

func parseNameFromPackage(bs []byte) (string, error) {
	p := &metaPkg{}
	err := yaml.Unmarshal(bs, p)
	return p.Metadata.Name, err
}

func parseKindFromPackage(bs []byte) (string, error) {
	p := &metaPkg{}
	err := yaml.Unmarshal(bs, p)
	fmt.Printf("Package: %v\n", p)
	return p.Metadata.Kind, err
}
