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

import (
	pkgmetav1 "github.com/yndd/ndd-core/apis/pkg/meta/v1"
	"github.com/yndd/ndd-core/internal/dag"
	nddv1 "github.com/yndd/ndd-runtime/apis/common/v1"
	"github.com/yndd/ndd-runtime/pkg/resource"
	"github.com/yndd/ndd-runtime/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RevisionActivationPolicy indicates how a package should activate its
// revisions.
type RevisionActivationPolicy string

var (
	// AutomaticActivation indicates that package should automatically activate
	// package revisions.
	AutomaticActivation RevisionActivationPolicy = "Automatic"
	// ManualActivation indicates that a user will manually activate package
	// revisions.
	ManualActivation RevisionActivationPolicy = "Manual"
)

// RefNames converts a slice of LocalObjectReferences to a slice of strings.
func RefNames(refs []corev1.LocalObjectReference) []string {
	stringRefs := make([]string, len(refs))
	for i, ref := range refs {
		stringRefs[i] = ref.Name
	}
	return stringRefs
}

var _ Package = &Provider{}

// +k8s:deepcopy-gen=false
type Package interface {
	resource.Object
	resource.Conditioned

	GetAutoPilot() bool
	SetAutoPilot(b bool)

	GetSource() string
	SetSource(s string)

	GetActivationPolicy() *RevisionActivationPolicy
	SetActivationPolicy(a *RevisionActivationPolicy)

	GetPackagePullSecrets() []corev1.LocalObjectReference
	SetPackagePullSecrets(s []corev1.LocalObjectReference)

	GetPackagePullPolicy() *corev1.PullPolicy
	SetPackagePullPolicy(i *corev1.PullPolicy)

	GetRevisionHistoryLimit() *int64
	SetRevisionHistoryLimit(l *int64)

	GetControllerConfigRef() *nddv1.Reference
	SetControllerConfigRef(r *nddv1.Reference)

	GetCurrentRevision() string
	SetCurrentRevision(r string)

	GetCurrentIdentifier() string
	SetCurrentIdentifier(r string)

	GetSkipDependencyResolution() *bool
	SetSkipDependencyResolution(*bool)
}

// GetCondition of this Provider.
func (p *Provider) GetCondition(ct nddv1.ConditionKind) nddv1.Condition {
	return p.Status.GetCondition(ct)
}

// SetConditions of this Provider.
func (p *Provider) SetConditions(c ...nddv1.Condition) {
	p.Status.SetConditions(c...)
}

// GetAutoPilot of this Provider.
func (p *Provider) GetAutoPilot() bool {
	return *p.Spec.AutoPilot
}

// SetAutoPilot of this Provider.
func (p *Provider) SetAutoPilot(b bool) {
	p.Spec.AutoPilot = utils.BoolPtr(b)
}

// GetSource of this Provider.
func (p *Provider) GetSource() string {
	return p.Spec.Package
}

// SetSource of this Provider.
func (p *Provider) SetSource(s string) {
	p.Spec.Package = s
}

// GetActivationPolicy of this Provider.
func (p *Provider) GetActivationPolicy() *RevisionActivationPolicy {
	return p.Spec.RevisionActivationPolicy
}

// SetActivationPolicy of this Provider.
func (p *Provider) SetActivationPolicy(a *RevisionActivationPolicy) {
	p.Spec.RevisionActivationPolicy = a
}

// GetPackagePullSecrets of this Provider.
func (p *Provider) GetPackagePullSecrets() []corev1.LocalObjectReference {
	return p.Spec.PackagePullSecrets
}

// SetPackagePullSecrets of this Provider.
func (p *Provider) SetPackagePullSecrets(s []corev1.LocalObjectReference) {
	p.Spec.PackagePullSecrets = s
}

// GetPackagePullPolicy of this Provider.
func (p *Provider) GetPackagePullPolicy() *corev1.PullPolicy {
	return p.Spec.PackagePullPolicy
}

// SetPackagePullPolicy of this Provider.
func (p *Provider) SetPackagePullPolicy(i *corev1.PullPolicy) {
	p.Spec.PackagePullPolicy = i
}

// GetRevisionHistoryLimit of this Provider.
func (p *Provider) GetRevisionHistoryLimit() *int64 {
	return p.Spec.RevisionHistoryLimit
}

// SetRevisionHistoryLimit of this Provider.
func (p *Provider) SetRevisionHistoryLimit(l *int64) {
	p.Spec.RevisionHistoryLimit = l
}

// GetControllerConfigRef of this Provider.
func (p *Provider) GetControllerConfigRef() *nddv1.Reference {
	return p.Spec.ControllerConfigReference
}

// SetControllerConfigRef of this Provider.
func (p *Provider) SetControllerConfigRef(r *nddv1.Reference) {
	p.Spec.ControllerConfigReference = r
}

// GetCurrentRevision of this Provider.
func (p *Provider) GetCurrentRevision() string {
	return p.Status.CurrentRevision
}

// SetCurrentRevision of this Provider.
func (p *Provider) SetCurrentRevision(s string) {
	p.Status.CurrentRevision = s
}

// GetSkipDependencyResolution of this Provider.
func (p *Provider) GetSkipDependencyResolution() *bool {
	return p.Spec.SkipDependencyResolution
}

// SetSkipDependencyResolution of this Provider.
func (p *Provider) SetSkipDependencyResolution(b *bool) {
	p.Spec.SkipDependencyResolution = b
}

// GetCurrentIdentifier of this Provider.
func (p *Provider) GetCurrentIdentifier() string {
	return p.Status.CurrentIdentifier
}

// SetCurrentIdentifier of this Provider.
func (p *Provider) SetCurrentIdentifier(s string) {
	p.Status.CurrentIdentifier = s
}

var _ PackageRevision = &ProviderRevision{}

// PackageRevision is the interface satisfied by package revision types.
// +k8s:deepcopy-gen=false
type PackageRevision interface {
	resource.Object
	resource.Conditioned

	GetObjects() []nddv1.TypedReference
	SetObjects(c []nddv1.TypedReference)

	GetControllerReference() nddv1.Reference
	SetControllerReference(c nddv1.Reference)

	GetSource() string
	SetSource(s string)

	GetPackagePullSecrets() []corev1.LocalObjectReference
	SetPackagePullSecrets(s []corev1.LocalObjectReference)

	GetPackagePullPolicy() *corev1.PullPolicy
	SetPackagePullPolicy(i *corev1.PullPolicy)

	GetDesiredState() PackageRevisionDesiredState
	SetDesiredState(d PackageRevisionDesiredState)

	GetControllerConfigRef() *nddv1.Reference
	SetControllerConfigRef(r *nddv1.Reference)

	GetRevision() int64
	SetRevision(r int64)

	GetAutoPilot() bool
	SetAutoPilot(b bool)

	GetSkipDependencyResolution() *bool
	SetSkipDependencyResolution(*bool)

	GetDependencyStatus() (found, installed, invalid int64)
	SetDependencyStatus(found, installed, invalid int64)
}

// GetCondition of this ProviderRevision.
func (p *ProviderRevision) GetCondition(ct nddv1.ConditionKind) nddv1.Condition {
	return p.Status.GetCondition(ct)
}

// SetConditions of this ProviderRevision.
func (p *ProviderRevision) SetConditions(c ...nddv1.Condition) {
	p.Status.SetConditions(c...)
}

// GetObjects of this ProviderRevision.
func (p *ProviderRevision) GetObjects() []nddv1.TypedReference {
	return p.Status.ObjectRefs
}

// SetObjects of this ProviderRevision.
func (p *ProviderRevision) SetObjects(c []nddv1.TypedReference) {
	p.Status.ObjectRefs = c
}

// GetControllerReference of this ProviderRevision.
func (p *ProviderRevision) GetControllerReference() nddv1.Reference {
	return p.Status.ControllerRef
}

// SetControllerReference of this ProviderRevision.
func (p *ProviderRevision) SetControllerReference(c nddv1.Reference) {
	p.Status.ControllerRef = c
}

// GetSource of this ProviderRevision.
func (p *ProviderRevision) GetSource() string {
	return p.Spec.PackageImage
}

// SetSource of this ProviderRevision.
func (p *ProviderRevision) SetSource(s string) {
	p.Spec.PackageImage = s
}

// GetPackagePullSecrets of this ProviderRevision.
func (p *ProviderRevision) GetPackagePullSecrets() []corev1.LocalObjectReference {
	return p.Spec.PackagePullSecrets
}

// SetPackagePullSecrets of this ProviderRevision.
func (p *ProviderRevision) SetPackagePullSecrets(s []corev1.LocalObjectReference) {
	p.Spec.PackagePullSecrets = s
}

// GetPackagePullPolicy of this ProviderRevision.
func (p *ProviderRevision) GetPackagePullPolicy() *corev1.PullPolicy {
	return p.Spec.PackagePullPolicy
}

// SetPackagePullPolicy of this ProviderRevision.
func (p *ProviderRevision) SetPackagePullPolicy(i *corev1.PullPolicy) {
	p.Spec.PackagePullPolicy = i
}

// GetDesiredState of this ProviderRevision.
func (p *ProviderRevision) GetDesiredState() PackageRevisionDesiredState {
	return p.Spec.DesiredState
}

// SetDesiredState of this ProviderRevision.
func (p *ProviderRevision) SetDesiredState(s PackageRevisionDesiredState) {
	p.Spec.DesiredState = s
}

// GetRevision of this ProviderRevision.
func (p *ProviderRevision) GetRevision() int64 {
	return p.Spec.Revision
}

// SetRevision of this ProviderRevision.
func (p *ProviderRevision) SetRevision(r int64) {
	p.Spec.Revision = r
}

// GetAutoPilot of this ProviderRevision.
func (p *ProviderRevision) GetAutoPilot() bool {
	return *p.Spec.AutoPilot
}

// SetAutoPilot of this ProviderRevision.
func (p *ProviderRevision) SetAutoPilot(b bool) {
	p.Spec.AutoPilot = utils.BoolPtr(b)
}

// GetDependencyStatus of this ProviderRevision.
func (p *ProviderRevision) GetDependencyStatus() (found, installed, invalid int64) {
	return p.Status.FoundDependencies, p.Status.InstalledDependencies, p.Status.InvalidDependencies
}

// SetDependencyStatus of this ProviderRevision.
func (p *ProviderRevision) SetDependencyStatus(found, installed, invalid int64) {
	p.Status.FoundDependencies = found
	p.Status.InstalledDependencies = installed
	p.Status.InvalidDependencies = invalid
}

// GetControllerConfigRef of this ProviderRevision.
func (p *ProviderRevision) GetControllerConfigRef() *nddv1.Reference {
	return p.Spec.ControllerConfigReference
}

// SetControllerConfigRef of this ProviderREvsion.
func (p *ProviderRevision) SetControllerConfigRef(r *nddv1.Reference) {
	p.Spec.ControllerConfigReference = r
}

// GetSkipDependencyResolution of this ProviderRevision.
func (p *ProviderRevision) GetSkipDependencyResolution() *bool {
	return p.Spec.SkipDependencyResolution
}

// SetSkipDependencyResolution of this ProviderRevision.
func (p *ProviderRevision) SetSkipDependencyResolution(b *bool) {
	p.Spec.SkipDependencyResolution = b
}

var _ PackageRevisionList = &ProviderRevisionList{}

// PackageRevisionList is the interface satisfied by package revision list
// types.
// +k8s:deepcopy-gen=false
type PackageRevisionList interface {
	client.ObjectList

	// GetRevisions gets the list of PackageRevisions in a PackageRevisionList.
	// This is a costly operation, but allows for treating different revision
	// list types as a single interface. If causing a performance bottleneck in
	// a shared reconciler, consider refactoring the controller to use a
	// reconciler for the specific type.
	GetRevisions() []PackageRevision
}

// GetRevisions of this ProviderRevisionList.
func (p *ProviderRevisionList) GetRevisions() []PackageRevision {
	prs := make([]PackageRevision, len(p.Items))
	for i, r := range p.Items {
		r := r // Pin range variable so we can take its address.
		prs[i] = &r
	}
	return prs
}

var _ dag.Node = &pkgmetav1.Dependency{}
var _ dag.Node = &LockPackage{}

// ToNodes converts LockPackages to DAG nodes.
func ToNodes(pkgs ...LockPackage) []dag.Node {
	nodes := make([]dag.Node, len(pkgs))
	for i, r := range pkgs {
		r := r // Pin range variable so we can take its address.
		nodes[i] = &r
	}
	return nodes
}

// Identifier returns the source of a LockPackage.
func (l *LockPackage) Identifier() string {
	return l.Source
}

// Neighbors returns dependencies of a LockPackage.
func (l *LockPackage) Neighbors() []dag.Node {
	nodes := make([]dag.Node, len(l.Dependencies))
	for i, r := range l.Dependencies {
		r := r // Pin range variable so we can take its address.
		nodes[i] = &r
	}
	return nodes
}

// AddNeighbors adds dependencies to a LockPackage. A LockPackage should always
// have all dependencies declared before being added to the Lock, so we no-op
// when adding a neighbor.
func (l *LockPackage) AddNeighbors(nodes ...dag.Node) error {
	return nil
}
