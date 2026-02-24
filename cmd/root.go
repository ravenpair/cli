package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ravenpair/cli/internal/adapters/http"
	"github.com/ravenpair/cli/internal/adapters/ws"
	"github.com/ravenpair/cli/internal/app"
)

var cfgFile string

// svc is the application service used by all sub-commands. It is wired with
// concrete adapters in PersistentPreRunE and can be replaced in tests.
var svc *app.Service

var rootCmd = &cobra.Command{
	Use:   "ravenpair",
	Short: "CLI tool to interact with the RavenPair server",
	Long: `ravenpair is a command-line interface for connecting to and managing
a RavenPair server via its REST API and WebSocket interface.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		serverURL := viper.GetString("server")
		token := viper.GetString("token")
		svc = app.New(http.New(serverURL, token), ws.New())
		return nil
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default $HOME/.ravenpair.yaml)")
	rootCmd.PersistentFlags().String("server", "http://localhost:8080", "RavenPair server URL")
	rootCmd.PersistentFlags().String("token", "", "authentication token")

	_ = viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))
	_ = viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ravenpair")
	}

	viper.SetEnvPrefix("RAVENPAIR")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
