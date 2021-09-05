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

package nddpkg

import (
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

const (
	errNoMatch    = "directory does not contain a compiled ndd package"
	errMultiMatch = "directory contains multiple compiled ndd packages"
)

// FindNddpkgInDir finds compiled Ndd packages in a directory.
func FindNddpkgInDir(fs afero.Fs, root string) (string, error) {
	f, err := fs.Open(root)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	files, err := f.Readdir(-1)
	if err != nil {
		return "", err
	}
	path := ""
	for _, file := range files {
		// Match only returns an error if NddpkgMatchPattern is malformed.
		match, _ := filepath.Match(NddpkgMatchPattern, file.Name()) //nolint:errcheck
		if !match {
			continue
		}
		if path != "" && match {
			return "", errors.New(errMultiMatch)
		}
		path = file.Name()
	}
	if path == "" {
		return "", errors.New(errNoMatch)
	}
	return path, nil
}
