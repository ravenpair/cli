package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(srv *httptest.Server) *Client {
	return New(srv.URL, "")
}

func TestGetStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/status" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	code, body, err := c.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus error: %v", err)
	}
	if code != http.StatusOK {
		t.Errorf("expected 200, got %d", code)
	}
	var v map[string]string
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if v["status"] != "ok" {
		t.Errorf("expected status ok, got %s", v["status"])
	}
}

func TestListPairs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/pairs" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":"1","name":"alpha"}]`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	code, body, err := c.ListPairs()
	if err != nil {
		t.Fatalf("ListPairs error: %v", err)
	}
	if code != http.StatusOK {
		t.Errorf("expected 200, got %d", code)
	}
	if len(body) == 0 {
		t.Error("expected non-empty body")
	}
}

func TestCreatePair(t *testing.T) {
	var gotName string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/pairs" {
			http.NotFound(w, r)
			return
		}
		var payload map[string]string
		_ = json.NewDecoder(r.Body).Decode(&payload)
		gotName = payload["name"]
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"99","name":"my-pair"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	code, _, err := c.CreatePair("my-pair")
	if err != nil {
		t.Fatalf("CreatePair error: %v", err)
	}
	if code != http.StatusCreated {
		t.Errorf("expected 201, got %d", code)
	}
	if gotName != "my-pair" {
		t.Errorf("expected name my-pair, got %s", gotName)
	}
}

func TestAuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "secret-token")
	_, _, err := c.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus error: %v", err)
	}
	if gotAuth != "Bearer secret-token" {
		t.Errorf("expected 'Bearer secret-token', got %q", gotAuth)
	}
}
