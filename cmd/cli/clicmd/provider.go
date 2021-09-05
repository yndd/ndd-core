package clicmd

import (
	"github.com/spf13/cobra"
)

// providerCmd represents the kubectl provider command
var providerCmd = &cobra.Command{
	Use:          "provider",
	Short:        "kubectl ndd provider cli",
	Long:         "kubectl ndd provider cli for usage with the network device driver in kubernetes",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(providerCmd)
}
