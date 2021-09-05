package clicmd

import (
	"os"

	nddv1 "github.com/yndd/ndd-core/apis/pkg/v1"
	nddpkg "github.com/yndd/ndd-core/internal/nddpkg"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

var (
	debug        bool
	buildChild   *BuildChild
	pushChild    *PushChild
	packageName  string
	providerName string
	scheme       = runtime.NewScheme()
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kubectl",
	Short: "A command line tool for interacting with ndd.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	//rootCmd.SilenceUsage = true
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug mode")
	//rootCmd.Flags().StringVarP(&packageName, "PackageName", "p", "", "Path to package. If not specified and only one package exists in current directory it will be used.")
	//rootCmd.Flags().StringVarP(&providerName, "providerName", "n", "", "Name of Provider.")

	buildChild = &BuildChild{
		fs:     afero.NewOsFs(),
		linter: nddpkg.NewProviderLinter(),
	}
	pushChild = &PushChild{
		fs: afero.NewOsFs(),
	}

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(nddv1.AddToScheme(scheme))
}
