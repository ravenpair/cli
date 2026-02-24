package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the current CLI version. It can be overridden at build time via
// -ldflags "-X github.com/ravenpair/cli/cmd.Version=x.y.z".
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of the ravenpair CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "ravenpair version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
