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
	"context"
	"fmt"
	"os"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	dvrv1 "github.com/yndd/ndd-core/apis/dvr/v1"
	metapkgv1 "github.com/yndd/ndd-core/apis/pkg/meta/v1"
	pkgv1 "github.com/yndd/ndd-core/apis/pkg/v1"
	"github.com/yndd/ndd-core/internal/initializer"
	"github.com/yndd/ndd-runtime/pkg/logging"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"github.com/spf13/cobra"
)

var (
	scheme = runtime.NewScheme()
	debug  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rbac",
	Short: "network device driver rbac",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	zlog := zap.New(zap.JSONEncoder())
	rootCmd.SilenceUsage = true
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug mode")

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(dvrv1.AddToScheme(scheme))
	utilruntime.Must(pkgv1.AddToScheme(scheme))
	utilruntime.Must(metapkgv1.AddToScheme(scheme))
	utilruntime.Must(extv1.AddToScheme(scheme))
	utilruntime.Must(extv1beta1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme

	cfg, err := ctrl.GetConfig()
	if err != nil {
		fmt.Printf("cannot get config %s\n", err)
	}

	cl, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		fmt.Printf("cannot create new kubernetes client %s\n", err)
	}
	i := initializer.New(cl,
		initializer.NewCRDWaiter([]string{
			fmt.Sprintf("%s.%s", "providerrevisions", pkgv1.Group),
		}, time.Minute, time.Second, logging.NewLogrLogger(zlog.WithName("nddrbacinit"))),
	)
	if err := i.Init(context.TODO()); err != nil {
		fmt.Printf("cannot initialize core %s\n", err)
	}
	fmt.Printf("Initialization has been completed\n")
}
