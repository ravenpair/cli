package http

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client is the HTTP adapter that implements ports.APIClient.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// New returns a new HTTP Client adapter.
func New(baseURL, token string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: &http.Client{},
	}
}

func (c *Client) do(method, path string, body io.Reader) (int, []byte, error) {
	url := c.baseURL + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return 0, nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("sending request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("reading response: %w", err)
	}

	return resp.StatusCode, data, nil
}

// GetStatus calls GET /api/status.
func (c *Client) GetStatus() (int, []byte, error) {
	return c.do(http.MethodGet, "/api/status", nil)
}

// ListPairs calls GET /api/pairs.
func (c *Client) ListPairs() (int, []byte, error) {
	return c.do(http.MethodGet, "/api/pairs", nil)
}

// CreatePair calls POST /api/pairs with an optional name payload.
func (c *Client) CreatePair(name string) (int, []byte, error) {
	var body io.Reader
	if name != "" {
		body = strings.NewReader(fmt.Sprintf(`{"name":%q}`, name))
	}
	return c.do(http.MethodPost, "/api/pairs", body)
}
