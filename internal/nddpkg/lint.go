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
	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	pkgmetav1 "github.com/yndd/ndd-core/apis/pkg/meta/v1"
	"github.com/yndd/ndd-core/internal/version"
	"github.com/yndd/ndd-runtime/pkg/parser"
)

const (
	errNotExactlyOneMeta  = "not exactly one package meta type"
	errNotMeta            = "meta type is not a package"
	errNotMetaProvider    = "package meta type is not a Provider"
	errNotMetaIntent      = "package meta type is not an Intent"
	errNotCRD             = "object is not a CRD"
	errBadConstraints     = "package version constraints are poorly formatted"
	errNddIncompatibleFmt = "package is not compatible with Ndd version (%s)"
)

// NewProviderLinter is a convenience function for creating a package linter for
// providers.
func NewProviderLinter() parser.Linter {
	return parser.NewPackageLinter(parser.PackageLinterFns(OneMeta), parser.ObjectLinterFns(IsProvider, PackageValidSemver), parser.ObjectLinterFns(IsCRD))
}

// NewIntentLinter is a convenience function for creating a package linter for
// intents.
func NewIntentLinter() parser.Linter {
	return parser.NewPackageLinter(parser.PackageLinterFns(OneMeta), parser.ObjectLinterFns(IsIntent, PackageValidSemver), parser.ObjectLinterFns(IsCRD))
}

// OneMeta checks that there is only one meta object in the package.
func OneMeta(pkg *parser.Package) error {
	if len(pkg.GetMeta()) != 1 {
		return errors.New(errNotExactlyOneMeta)
	}
	return nil
}

// IsProvider checks that an object is a Provider meta type.
func IsProvider(o runtime.Object) error {
	po, _ := TryConvert(o, &pkgmetav1.Provider{})
	if _, ok := po.(*pkgmetav1.Provider); !ok {
		return errors.New(errNotMetaProvider)
	}
	return nil
}

// IsIntent checks that an object is a Intent meta type.
func IsIntent(o runtime.Object) error {
	po, _ := TryConvert(o, &pkgmetav1.Intent{})
	if _, ok := po.(*pkgmetav1.Intent); !ok {
		return errors.New(errNotMetaIntent)
	}
	return nil
}

// PackageNddCompatible checks that the current Ndd version is
// compatible with the package constraints.
func PackageNddCompatible(v version.Operations) parser.ObjectLinterFn {
	return func(o runtime.Object) error {
		p, ok := TryConvertToPkg(o, &pkgmetav1.Provider{})
		if !ok {
			return errors.New(errNotMeta)
		}

		if p.GetNddConstraints() == nil {
			return nil
		}
		in, err := v.InConstraints(p.GetNddConstraints().Version)
		if err != nil {
			return errors.Wrapf(err, errNddIncompatibleFmt, v.GetVersionString())
		}
		if !in {
			return errors.Errorf(errNddIncompatibleFmt, v.GetVersionString())
		}
		return nil
	}
}

// PackageValidSemver checks that the package uses valid semver ranges.
func PackageValidSemver(o runtime.Object) error {
	p, ok := TryConvertToPkg(o, &pkgmetav1.Provider{})
	if !ok {
		return errors.New(errNotMeta)
	}

	if p.GetNddConstraints() == nil {
		return nil
	}
	if _, err := semver.NewConstraint(p.GetNddConstraints().Version); err != nil {
		return errors.Wrap(err, errBadConstraints)
	}
	return nil
}

// IsCRD checks that an object is a CustomResourceDefinition.
func IsCRD(o runtime.Object) error {
	switch o.(type) {
	case *extv1beta1.CustomResourceDefinition, *extv1.CustomResourceDefinition:
		return nil
	default:
		return errors.New(errNotCRD)
	}
}
