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

package kubectlnddcmd

import (
	"github.com/spf13/afero"
	"github.com/yndd/ndd-runtime/pkg/parser"
)

type BuildChild struct {
	name           string
	providerLinter parser.Linter
	intentLinter   parser.Linter
	fs             afero.Fs
}

type PushChild struct {
	fs afero.Fs
}

/*
// pushProviderCmd pushes a Provider.
type PushProviderCmd struct {
	Tag string `arg:"" help:"Tag of the package to be pushed. Must be a valid OCI image tag."`
}

type PushIntentCmd struct {
	Tag string `arg:"" help:"Tag of the package to be pushed. Must be a valid OCI image tag."`
}
*/
