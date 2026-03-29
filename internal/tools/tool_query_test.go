package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestQueryRouter_Flows(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/devices":
			w.Write([]byte(`{"devices":[{"device_name":"r1"}]}`))
		case "/query/topXdata":
			w.Write([]byte(`{"results":[{"data":[{"key":"AS15169","avg_bits_per_sec":1000000}]}]}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeQueryRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"action":    "flows",
		"metric":    "bytes",
		"dimension": "AS_dst",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "AS15169") {
		t.Errorf("expected AS15169 in response, got: %s", text)
	}
}

func TestQueryRouter_TimeSeries(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/devices":
			w.Write([]byte(`{"devices":[{"device_name":"r1"}]}`))
		case "/query/topXdata":
			w.Write([]byte(`{"results":[{"data":[{"key":"site1","avg_bits_per_sec":500000,"p95th_bits_per_sec":800000,"max_bits_per_sec":1000000}]}]}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeQueryRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"action":    "time_series",
		"metric":    "bytes",
		"dimension": "i_device_site_name",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
}

func TestQueryRouter_TopDDoS(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/alerts-active/alerts-history") {
			w.Write([]byte(`[{"alarm_start":"2026-03-25T13:14:07Z","alarm_end":"2026-03-25T13:24:07Z","policy_name":"DDoS","alert_key":"1.2.3.4","alert_severity":"critical","alert_value":5000000000,"alert_metric":["bits","packets"]}]`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeQueryRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"action":  "top_ddos",
		"sort_by": "bps",
		"top":     5,
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
}

func TestQueryRouter_UnknownAction(t *testing.T) {
	handler := makeQueryRouter(testClientWithServer("http://unused"))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"action": "bogus"}

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

func TestQueryRouter_MissingAction(t *testing.T) {
	handler := makeQueryRouter(testClientWithServer("http://unused"))
	req := mcp.CallToolRequest{}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error when action is missing")
	}
}
