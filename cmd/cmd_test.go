package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/ravenpair/cli/internal/app"
	"github.com/ravenpair/cli/internal/ports"
)

// --- mock implementations of ports ---

type mockAPIClient struct {
	getStatusFn  func() (int, []byte, error)
	listPairsFn  func() (int, []byte, error)
	createPairFn func(string) (int, []byte, error)
}

func (m *mockAPIClient) GetStatus() (int, []byte, error)          { return m.getStatusFn() }
func (m *mockAPIClient) ListPairs() (int, []byte, error)          { return m.listPairsFn() }
func (m *mockAPIClient) CreatePair(name string) (int, []byte, error) { return m.createPairFn(name) }

type mockWSClient struct {
	dialFn func(ctx context.Context, wsURL string, headers map[string]string, onMessage ports.MessageHandler) error
}

func (m *mockWSClient) Dial(ctx context.Context, wsURL string, headers map[string]string, onMessage ports.MessageHandler) error {
	return m.dialFn(ctx, wsURL, headers, onMessage)
}

// setSvc replaces the package-level service with one backed by the given mocks.
func setSvc(api ports.APIClient, ws ports.WSClient) {
	svc = app.New(api, ws)
}

// --- tests ---

func TestVersionCmd(t *testing.T) {
	Version = "1.2.3"

	buf := new(bytes.Buffer)
	versionCmd.SetOut(buf)
	versionCmd.SetErr(new(bytes.Buffer))

	versionCmd.Run(versionCmd, nil)

	got := buf.String()
	if !strings.Contains(got, "1.2.3") {
		t.Errorf("expected version 1.2.3 in output, got: %s", got)
	}
}

func TestStatusCmd(t *testing.T) {
	setSvc(&mockAPIClient{
		getStatusFn: func() (int, []byte, error) {
			return 200, []byte(`{"status":"ok"}`), nil
		},
	}, nil)

	buf := new(bytes.Buffer)
	statusCmd.SetOut(buf)
	statusCmd.SetErr(new(bytes.Buffer))

	if err := statusCmd.RunE(statusCmd, nil); err != nil {
		t.Fatalf("status command failed: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "200") {
		t.Errorf("expected HTTP 200 in output, got: %s", got)
	}
	if !strings.Contains(got, "ok") {
		t.Errorf("expected 'ok' in response body, got: %s", got)
	}
}

func TestListCmd(t *testing.T) {
	setSvc(&mockAPIClient{
		listPairsFn: func() (int, []byte, error) {
			return 200, []byte(`[{"id":"1","name":"alpha"},{"id":"2","name":"beta"}]`), nil
		},
	}, nil)

	buf := new(bytes.Buffer)
	listCmd.SetOut(buf)
	listCmd.SetErr(new(bytes.Buffer))

	if err := listCmd.RunE(listCmd, nil); err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "alpha") || !strings.Contains(got, "beta") {
		t.Errorf("expected pair names in output, got: %s", got)
	}
}

func TestPairCmd(t *testing.T) {
	var gotName string
	setSvc(&mockAPIClient{
		createPairFn: func(name string) (int, []byte, error) {
			gotName = name
			return 201, []byte(`{"id":"42","name":"test-pair"}`), nil
		},
	}, nil)

	buf := new(bytes.Buffer)
	pairCmd.ResetFlags()
	pairCmd.Flags().String("name", "test-pair", "name for the pair session")
	pairCmd.SetOut(buf)
	pairCmd.SetErr(new(bytes.Buffer))

	if err := pairCmd.RunE(pairCmd, nil); err != nil {
		t.Fatalf("pair command failed: %v", err)
	}

	if gotName != "test-pair" {
		t.Errorf("expected name 'test-pair', got %q", gotName)
	}
	got := buf.String()
	if !strings.Contains(got, "201") {
		t.Errorf("expected HTTP 201 in output, got: %s", got)
	}
	if !strings.Contains(got, "42") {
		t.Errorf("expected id 42 in response, got: %s", got)
	}
}

func TestConnectCmd(t *testing.T) {
	setSvc(nil, &mockWSClient{
		dialFn: func(ctx context.Context, _ string, _ map[string]string, onMessage ports.MessageHandler) error {
			// Simulate two text messages then server close
			onMessage(1, []byte("hello"))
			onMessage(1, []byte("world"))
			return nil // server-side close
		},
	})

	buf := new(bytes.Buffer)
	connectCmd.ResetFlags()
	connectCmd.Flags().String("path", "/ws", "WebSocket endpoint path")
	connectCmd.SetOut(buf)
	connectCmd.SetErr(new(bytes.Buffer))

	if err := connectCmd.RunE(connectCmd, nil); err != nil {
		t.Fatalf("connect command failed: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "hello") || !strings.Contains(got, "world") {
		t.Errorf("expected messages in output, got: %s", got)
	}
}
