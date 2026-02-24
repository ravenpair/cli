package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Open a WebSocket connection to the RavenPair server",
	Long: `Establish a persistent WebSocket connection to the RavenPair server.
Messages received from the server are printed to stdout.
Press Ctrl+C to close the connection.`,
	RunE: runConnect,
}

func init() {
	rootCmd.AddCommand(connectCmd)
	connectCmd.Flags().String("path", "/ws", "WebSocket endpoint path")
}

func runConnect(cmd *cobra.Command, args []string) error {
	path, _ := cmd.Flags().GetString("path")
	serverURL := viper.GetString("server")
	token := viper.GetString("token")

	fmt.Fprintf(cmd.OutOrStdout(), "Connecting to %s%s\n", serverURL, path)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer signal.Stop(interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	fmt.Fprintln(cmd.OutOrStdout(), "Connected. Press Ctrl+C to disconnect.")

	connDone := make(chan error, 1)
	go func() {
		connDone <- svc.Connect(ctx, serverURL, path, token, func(msgType int, data []byte) {
			switch msgType {
			case websocket.TextMessage:
				fmt.Fprintf(cmd.OutOrStdout(), "< %s\n", data)
			case websocket.BinaryMessage:
				fmt.Fprintf(cmd.OutOrStdout(), "< [binary %d bytes]\n", len(data))
			}
		})
	}()

	select {
	case err := <-connDone:
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "connection error: %v\n", err)
			cancel()
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Connection closed by server.")
		cancel()
	case <-interrupt:
		fmt.Fprintln(cmd.OutOrStdout(), "\nInterrupted. Closing connection...")
		cancel()
		<-connDone
	}

	return nil
}
