package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

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

	// Convert http(s) scheme to ws(s)
	wsURL := toWebSocketURL(serverURL) + path
	fmt.Fprintf(cmd.OutOrStdout(), "Connecting to %s\n", wsURL)

	headers := http.Header{}
	if token := viper.GetString("token"); token != "" {
		headers.Set("Authorization", "Bearer "+token)
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(wsURL, headers)
	if err != nil {
		return fmt.Errorf("WebSocket dial: %w", err)
	}
	defer conn.Close()

	fmt.Fprintln(cmd.OutOrStdout(), "Connected. Press Ctrl+C to disconnect.")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})

	// Read messages from server
	go func() {
		defer close(done)
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					fmt.Fprintf(cmd.ErrOrStderr(), "read error: %v\n", err)
				}
				return
			}
			switch msgType {
			case websocket.TextMessage:
				fmt.Fprintf(cmd.OutOrStdout(), "< %s\n", msg)
			case websocket.BinaryMessage:
				fmt.Fprintf(cmd.OutOrStdout(), "< [binary %d bytes]\n", len(msg))
			}
		}
	}()

	select {
	case <-done:
		fmt.Fprintln(cmd.OutOrStdout(), "Connection closed by server.")
	case <-interrupt:
		fmt.Fprintln(cmd.OutOrStdout(), "\nInterrupted. Closing connection...")
		err = conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)
		if err != nil {
			return fmt.Errorf("sending close: %w", err)
		}
		select {
		case <-done:
		case <-time.After(3 * time.Second):
		}
	}

	return nil
}

// toWebSocketURL converts an http/https URL to a ws/wss URL.
func toWebSocketURL(u string) string {
	u = strings.TrimRight(u, "/")
	switch {
	case strings.HasPrefix(u, "https://"):
		return "wss://" + strings.TrimPrefix(u, "https://")
	case strings.HasPrefix(u, "http://"):
		return "ws://" + strings.TrimPrefix(u, "http://")
	default:
		return u
	}
}
