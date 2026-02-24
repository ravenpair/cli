package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

func newHTTPClient() *http.Client {
	return &http.Client{}
}

func doRequest(method, path string, body io.Reader) ([]byte, int, error) {
	serverURL := strings.TrimRight(viper.GetString("server"), "/")
	url := serverURL + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if token := viper.GetString("token"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := newHTTPClient().Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("sending request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response: %w", err)
	}

	return data, resp.StatusCode, nil
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
	data, status, err := doRequest(http.MethodGet, "/api/status", nil)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "HTTP %d\n", status)
	printJSON(out, data)
	return nil
}

func runPair(cmd *cobra.Command, args []string) error {
	name, _ := cmd.Flags().GetString("name")

	var body io.Reader
	if name != "" {
		payload := fmt.Sprintf(`{"name":%q}`, name)
		body = strings.NewReader(payload)
	}

	data, status, err := doRequest(http.MethodPost, "/api/pairs", body)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "HTTP %d\n", status)
	printJSON(out, data)
	return nil
}

func runList(cmd *cobra.Command, args []string) error {
	data, status, err := doRequest(http.MethodGet, "/api/pairs", nil)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "HTTP %d\n", status)
	printJSON(out, data)
	return nil
}
