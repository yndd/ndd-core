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

package manager

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"
	v1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/yndd/ndd-core/internal/nddpkg"
	"github.com/yndd/ndd-runtime/pkg/logging"
)

const (
	errFetchPackage = "failed to fetch package digest from remote"
)

// Revisioner extracts a revision name for a package source.
type Revisioner interface {
	Revision(context.Context, logging.Logger, v1.Package) (string, error)
}

// PackageRevisioner extracts a revision name for a package source.
type PackageRevisioner struct {
	fetcher nddpkg.Fetcher
}

// NewPackageRevisioner returns a new PackageRevisioner.
func NewPackageRevisioner(fetcher nddpkg.Fetcher) *PackageRevisioner {
	return &PackageRevisioner{
		fetcher: fetcher,
	}
}

// Revision extracts a revision name for a package source.
func (r *PackageRevisioner) Revision(ctx context.Context, log logging.Logger, p v1.Package) (string, error) {
	pullPolicy := p.GetPackagePullPolicy()
	if pullPolicy != nil && *pullPolicy == corev1.PullNever {
		return nddpkg.FriendlyID(p.GetName(), p.GetSource()), nil
	}
	if pullPolicy != nil && *pullPolicy == corev1.PullIfNotPresent {
		if p.GetCurrentIdentifier() == p.GetSource() {
			return p.GetCurrentRevision(), nil
		}
	}
	ref, err := name.ParseReference(p.GetSource())
	if err != nil {
		return "", err
	}
	log.Debug("Head fetcher", "Source", p.GetSource(), "CurrentIdentifier", p.GetCurrentIdentifier())
	d, err := r.fetcher.Head(ctx, ref, v1.RefNames(p.GetPackagePullSecrets())...)
	if err != nil || d == nil {
		return "", errors.Wrap(err, errFetchPackage)
	}
	return nddpkg.FriendlyID(p.GetName(), d.Digest.Hex), nil
}

// NopRevisioner returns an empty revision name.
type NopRevisioner struct{}

// NewNopRevisioner creates a NopRevisioner.
func NewNopRevisioner() *NopRevisioner {
	return &NopRevisioner{}
}

// Revision returns an empty revision name and no error.
func (d *NopRevisioner) Revision(context.Context, logging.Logger, v1.Package) (string, error) {
	return "", nil
}
