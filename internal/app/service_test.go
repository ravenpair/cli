package app

import (
	"context"
	"testing"

	"github.com/ravenpair/cli/internal/ports"
)

func TestToWebSocketURL(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"http://localhost:8080", "ws://localhost:8080"},
		{"https://example.com", "wss://example.com"},
		{"http://example.com/", "ws://example.com"},
		{"ws://already-ws.com", "ws://already-ws.com"},
	}
	for _, tc := range cases {
		got := toWebSocketURL(tc.input)
		if got != tc.want {
			t.Errorf("toWebSocketURL(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestServiceConnectBuildsURL(t *testing.T) {
	var capturedURL string
	mock := &mockWSClient{
		dialFn: func(_ context.Context, wsURL string, _ map[string]string, _ ports.MessageHandler) error {
			capturedURL = wsURL
			return nil
		},
	}
	svc := New(nil, mock)
	_ = svc.Connect(context.Background(), "http://localhost:8080", "/ws", "", func(int, []byte) {})
	if capturedURL != "ws://localhost:8080/ws" {
		t.Errorf("unexpected wsURL: %s", capturedURL)
	}
}

func TestServiceConnectSetsToken(t *testing.T) {
	var capturedHeaders map[string]string
	mock := &mockWSClient{
		dialFn: func(_ context.Context, _ string, headers map[string]string, _ ports.MessageHandler) error {
			capturedHeaders = headers
			return nil
		},
	}
	svc := New(nil, mock)
	_ = svc.Connect(context.Background(), "http://localhost:8080", "/ws", "tok123", func(int, []byte) {})
	if capturedHeaders["Authorization"] != "Bearer tok123" {
		t.Errorf("expected Bearer tok123, got %s", capturedHeaders["Authorization"])
	}
}

// mockWSClient is a test double for ports.WSClient.
type mockWSClient struct {
	dialFn func(ctx context.Context, wsURL string, headers map[string]string, onMessage ports.MessageHandler) error
}

func (m *mockWSClient) Dial(ctx context.Context, wsURL string, headers map[string]string, onMessage ports.MessageHandler) error {
	return m.dialFn(ctx, wsURL, headers, onMessage)
}
