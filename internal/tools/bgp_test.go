package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestListBGPMonitorsHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bgp_monitoring/v202205beta1/monitors" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Write([]byte(`{"monitors":[{"id":"1","name":"test-monitor"}]}`))
	}))
	defer srv.Close()

	handler := makeListBGPMonitorsHandler(testClientWithServer(srv.URL))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "test-monitor") {
		t.Errorf("expected test-monitor in response, got: %s", text)
	}
}

func TestGetBGPRoutesHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bgp_monitoring/v202205beta1/monitors/42/routes" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Write([]byte(`{"routes":[{"prefix":"8.8.8.0/24","as_path":"15169"}]}`))
	}))
	defer srv.Close()

	handler := makeGetBGPRoutesHandler(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"monitor_id": "42",
		"prefix":     "8.8.8.0/24",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "8.8.8.0/24") {
		t.Errorf("expected prefix in response, got: %s", text)
	}
}

func TestGetBGPRoutesHandler_MissingMonitorID(t *testing.T) {
	handler := makeGetBGPRoutesHandler(testClientWithServer("http://unused"))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result when monitor_id missing")
	}
}
