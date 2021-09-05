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

package rbac

import (
	"github.com/yndd/ndd-core/internal/controllers/rbac/binding"
	"github.com/yndd/ndd-core/internal/controllers/rbac/roles"
	"github.com/yndd/ndd-runtime/pkg/logging"
	ctrl "sigs.k8s.io/controller-runtime"
)

// The ManagementPolicy specifies which roles the RBAC manager should manage.
type ManagementPolicy string

const (
	// ManagementPolicyAll indicates that all RBAC manager functionality should
	// be enabled.
	ManagementPolicyAll ManagementPolicy = "All"

	// ManagementPolicyBasic indicates that basic RBAC manager functionality
	// should be enabled. The RBAC manager will create ClusterRoles for each
	// XRD. The ClusterRoles it creates will aggregate to the core Crossplane
	// ClusterRoles (e.g. crossplane, crossplane-admin, etc).
	ManagementPolicyBasic ManagementPolicy = "Basic"
)

// Setup RBAC manager controllers.
func Setup(mgr ctrl.Manager, l logging.Logger, mp ManagementPolicy, allowClusterRole string) error {
	// Basic controllers.
	fns := []func(ctrl.Manager, logging.Logger) error{
		//definition.Setup,
		binding.Setup,
	}

	//if mp == ManagementPolicyAll {
	//	fns = append(fns, namespace.Setup)
	//}

	for _, setup := range fns {
		if err := setup(mgr, l); err != nil {
			return err
		}
	}

	return roles.Setup(mgr, l, allowClusterRole)
}
