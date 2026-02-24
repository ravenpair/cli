package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Interact with the RavenPair REST API",
	Long:  `Send requests to the RavenPair server REST API endpoints.`,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get server status",
	Long:  `Retrieve the health/status of the RavenPair server.`,
	RunE:  runStatus,
}

var pairCmd = &cobra.Command{
	Use:   "pair",
	Short: "Create or get a pair",
	Long:  `Create a new pair session or retrieve an existing one from the RavenPair server.`,
	RunE:  runPair,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List active pairs",
	Long:  `List all currently active pair sessions on the RavenPair server.`,
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(apiCmd)
	apiCmd.AddCommand(statusCmd)
	apiCmd.AddCommand(pairCmd)
	apiCmd.AddCommand(listCmd)

	pairCmd.Flags().String("name", "", "name for the pair session")
}

func printJSON(w io.Writer, data []byte) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Fprintln(w, string(data))
		return
	}
	pretty, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintln(w, string(data))
		return
	}
	fmt.Fprintln(w, string(pretty))
}

func runStatus(cmd *cobra.Command, args []string) error {
	statusCode, body, err := svc.API.GetStatus()
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "HTTP %d\n", statusCode)
	printJSON(out, body)
	return nil
}

func runPair(cmd *cobra.Command, args []string) error {
	name, _ := cmd.Flags().GetString("name")
	statusCode, body, err := svc.API.CreatePair(name)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "HTTP %d\n", statusCode)
	printJSON(out, body)
	return nil
}

func runList(cmd *cobra.Command, args []string) error {
	statusCode, body, err := svc.API.ListPairs()
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "HTTP %d\n", statusCode)
	printJSON(out, body)
	return nil
}
