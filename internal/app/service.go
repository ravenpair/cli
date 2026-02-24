package app

import (
	"context"
	"strings"

	"github.com/ravenpair/cli/internal/ports"
)

// Service is the application layer that orchestrates the outgoing ports.
type Service struct {
	API ports.APIClient
	WS  ports.WSClient
}

// New creates a new Service wiring together the given port implementations.
func New(api ports.APIClient, ws ports.WSClient) *Service {
	return &Service{API: api, WS: ws}
}

// Connect builds the WebSocket URL from serverURL + path, attaches the Bearer
// token header when provided, and delegates to the WSClient port.
func (s *Service) Connect(ctx context.Context, serverURL, path, token string, onMessage ports.MessageHandler) error {
	wsURL := toWebSocketURL(serverURL) + path
	headers := map[string]string{}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	return s.WS.Dial(ctx, wsURL, headers, onMessage)
}

// toWebSocketURL converts an http(s):// URL to ws(s)://.
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
