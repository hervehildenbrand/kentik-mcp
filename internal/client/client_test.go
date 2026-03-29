package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hervehildenbrand/kentik-mcp/internal/config"
)

func testClient(url string) *Client {
	return New(&config.Config{
		Email:    "test@example.com",
		APIToken: "test-token",
		Region:   "EU",
		V5Base:   url,
		V6Base:   url,
	})
}

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Email:    "test@example.com",
		APIToken: "token123",
		Region:   "EU",
		V5Base:   "https://api.kentik.eu/api/v5",
		V6Base:   "https://grpc.api.kentik.eu",
	}
	c := New(cfg)
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestV5_SetsAuthHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-CH-Auth-Email"); got != "test@example.com" {
			t.Errorf("expected email header test@example.com, got %s", got)
		}
		if got := r.Header.Get("X-CH-Auth-API-Token"); got != "test-token" {
			t.Errorf("expected token header test-token, got %s", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("expected content-type application/json, got %s", got)
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := testClient(srv.URL)
	data, err := c.V5("GET", "/devices", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	json.Unmarshal(data, &result)
	if result["ok"] != true {
		t.Error("expected ok:true in response")
	}
}

func TestV5_ReturnsErrorOnNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer srv.Close()

	c := testClient(srv.URL)
	_, err := c.V5("GET", "/devices", nil)
	if err == nil {
		t.Error("expected error on 401 response")
	}
}

func TestV5_SendsJSONBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["key"] != "value" {
			t.Errorf("expected body key=value, got %v", body)
		}
		w.Write([]byte(`{"result":"ok"}`))
	}))
	defer srv.Close()

	c := testClient(srv.URL)
	_, err := c.V5("POST", "/query", map[string]any{"key": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestV6_UsesV6Base(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bgp_monitoring/v202205beta1/monitors" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"monitors":[]}`))
	}))
	defer srv.Close()

	c := testClient(srv.URL)
	_, err := c.V6("GET", "/bgp_monitoring/v202205beta1/monitors", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
