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

package v1

import "github.com/yndd/ndd-core/internal/dag"

var _ Pkg = &Provider{}

// Pkg is a description of a Ndd package.
// +k8s:deepcopy-gen=false
type Pkg interface {
	GetNddConstraints() *NddConstraints
	GetDependencies() []Dependency
}

// GetNddConstraints gets the Provider package's ndd version
// constraints.
func (c *Provider) GetNddConstraints() *NddConstraints {
	return c.Spec.MetaSpec.Ndd
}

// GetDependencies gets the Provider package's dependencies.
func (c *Provider) GetDependencies() []Dependency {
	return c.Spec.MetaSpec.DependsOn
}

// Identifier returns a dependency's source.
func (d *Dependency) Identifier() string {
	return d.Package
}

// Neighbors in is a no-op for dependencies because we are not yet aware of its
// dependencies.
func (d *Dependency) Neighbors() []dag.Node {
	return nil
}

// AddNeighbors is a no-op for dependencies. We should never be adding neighbors
// to a dependency.
func (d *Dependency) AddNeighbors(...dag.Node) error {
	return nil
}
