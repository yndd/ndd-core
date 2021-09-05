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

package rbaccmd

import (
	"time"

	"github.com/yndd/ndd-core/internal/controllers/rbac"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// Available RBAC management policies.
const (
	ManagementPolicyAll   = string(rbac.ManagementPolicyAll)
	ManagementPolicyBasic = string(rbac.ManagementPolicyBasic)
)

var (
	metricsAddr          string
	probeAddr            string
	enableLeaderElection bool
	concurrency          int
	sync                 string
	managementPolicy     string
	providerClusterRole  string
)

// startCmd represents the start command for the network device driver
var startCmd = &cobra.Command{
	Use:          "start",
	Short:        "start the network device driver rbac",
	Long:         "start the network device driver rbac",
	Aliases:      []string{"start"},
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		zlog := zap.New(zap.UseDevMode(debug), zap.JSONEncoder())
		if debug {
			// Only use a logr.Logger when debug is on
			ctrl.SetLogger(zlog)
		}
		zlog.Info("create ndd rbac manager")
		syncPeriod, err := time.ParseDuration(sync)
		if err != nil {
			return errors.Wrap(err, "Cannot parse sync duration")
		}
		mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
			Scheme:                 scheme,
			MetricsBindAddress:     metricsAddr,
			Port:                   9443,
			HealthProbeBindAddress: probeAddr,
			LeaderElection:         enableLeaderElection,
			LeaderElectionID:       "c66ce353.rbac.ndd.yndd.io",
			SyncPeriod:             &syncPeriod,
		})
		if err != nil {
			return errors.Wrap(err, "Cannot create manager")
		}

		if err := rbac.Setup(mgr, logging.NewLogrLogger(zlog.WithName("nddrbac")), rbac.ManagementPolicy(managementPolicy), providerClusterRole); err != nil {
			return errors.Wrap(err, "Cannot add rbac controllers to manager")
		}

		// +kubebuilder:scaffold:builder

		if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
			return errors.Wrap(err, "unable to set up health check")
		}
		if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
			return errors.Wrap(err, "unable to set up ready check")
		}

		zlog.Info("starting nddrbac manager")
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			return errors.Wrap(err, "problem running manager")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVarP(&metricsAddr, "metrics-bind-address", "m", ":8080", "The address the metric endpoint binds to.")
	startCmd.Flags().StringVarP(&probeAddr, "health-probe-bind-address", "p", ":8081", "The address the probe endpoint binds to.")
	startCmd.Flags().BoolVarP(&enableLeaderElection, "leader-elect", "l", false, "Enable leader election for controller manager. "+
		"Enabling this will ensure there is only one active controller manager.")
	startCmd.Flags().IntVarP(&concurrency, "concurrency", "", 1, "Number of items to process simultaneously")
	startCmd.Flags().StringVarP(&sync, "sync", "s", "1h", "Controller manager sync period duration such as 300ms, 1.5h or 2h45m")
	startCmd.Flags().StringVarP(&managementPolicy, "management-policy", "", ManagementPolicyBasic, "RBAC management policy.")
	startCmd.Flags().StringVarP(&providerClusterRole, "provider-clusterrole", "c", "", "A ClusterRole enumerating the permissions provider packages may request.")

}

func nddConcurrency(c int) controller.Options {
	return controller.Options{MaxConcurrentReconciles: c}
}
