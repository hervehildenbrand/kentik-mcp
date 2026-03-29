package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestBGPRouter_ListMonitors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/bgp_monitoring/v202205beta1/monitors" {
			w.Write([]byte(`{"monitors":[{"id":"1","name":"Google Prefix Monitor"}]}`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeBGPRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"action": "list_monitors"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "Google Prefix Monitor") {
		t.Errorf("expected monitor name in response, got: %s", text)
	}
}

func TestBGPRouter_GetRoutes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/bgp_monitoring/v202205beta1/monitors/42/routes" && r.Method == "POST" {
			w.Write([]byte(`{"routes":[{"prefix":"8.8.8.0/24","asPath":["15169"]}]}`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeBGPRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"action":     "get_routes",
		"monitor_id": "42",
		"prefix":     "8.8.8.0/24",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
}

func TestBGPRouter_UnknownAction(t *testing.T) {
	handler := makeBGPRouter(testClientWithServer("http://unused"))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"action": "invalid"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for unknown action")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "Unknown action") {
		t.Errorf("expected 'Unknown action' error, got: %s", text)
	}
}
