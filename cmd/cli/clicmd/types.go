package clicmd

import (
	"github.com/spf13/afero"
	"github.com/yndd/ndd-runtime/pkg/parser"
)

type BuildChild struct {
	name   string
	linter parser.Linter
	fs     afero.Fs
}

type PushChild struct {
	fs afero.Fs
}

// pushProviderCmd pushes a Provider.
type PushProviderCmd struct {
	Tag string `arg:"" help:"Tag of the package to be pushed. Must be a valid OCI image tag."`
}
