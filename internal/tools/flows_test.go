package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestParseDimensions(t *testing.T) {
	got := parseDimensions("AS_src, AS_dst")
	if len(got) != 2 || got[0] != "AS_src" || got[1] != "AS_dst" {
		t.Errorf("parseDimensions = %v, want [AS_src, AS_dst]", got)
	}
}

func TestParseDimensions_Single(t *testing.T) {
	got := parseDimensions("IP_dst")
	if len(got) != 1 || got[0] != "IP_dst" {
		t.Errorf("parseDimensions = %v, want [IP_dst]", got)
	}
}

func TestGetOutsort(t *testing.T) {
	tests := []struct {
		metric string
		want   string
	}{
		{"bytes", "avg_bits_per_sec"},
		{"in_bytes", "avg_bits_per_sec"},
		{"packets", "avg_pkts_per_sec"},
		{"fps", "avg_flows_per_sec"},
	}
	for _, tt := range tests {
		if got := getOutsort(tt.metric); got != tt.want {
			t.Errorf("getOutsort(%q) = %q, want %q", tt.metric, got, tt.want)
		}
	}
}

// flowMockServer handles both /devices and flow query endpoints
func flowMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/devices":
			w.Write([]byte(`{"devices":[{"device_name":"router1"},{"device_name":"router2"}]}`))
		case "/query/topXdata":
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			queries, ok := body["queries"].([]any)
			if !ok || len(queries) == 0 {
				t.Error("expected queries array in body")
			}
			w.Write([]byte(`{"results":[{"data":[{"key":"AS15169","avg_bits_per_sec":1000000,"p95th_bits_per_sec":1500000,"max_bits_per_sec":2000000}]}]}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
}

func TestQueryFlowsHandler(t *testing.T) {
	srv := flowMockServer(t)
	defer srv.Close()

	handler := makeQueryFlowsHandler(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"metric":    "bytes",
		"dimension": "AS_dst",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success result, got error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "AS15169") {
		t.Errorf("expected AS15169 in response, got: %s", text)
	}
	if !strings.Contains(text, "Flow Query Results") {
		t.Errorf("expected markdown summary, got: %s", text)
	}
}

func TestQueryFlowsHandler_WithDeviceName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/query/topXdata" {
			w.Write([]byte(`{"results":[{"data":[{"key":"AS15169","avg_bits_per_sec":500000}]}]}`))
		} else {
			t.Errorf("should not call %s when device_name is provided", r.URL.Path)
		}
	}))
	defer srv.Close()

	handler := makeQueryFlowsHandler(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"metric":      "bytes",
		"dimension":   "AS_dst",
		"device_name": "router1",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success result, got error: %v", result.Content)
	}
}

func TestQueryTimeSeriesHandler(t *testing.T) {
	srv := flowMockServer(t)
	defer srv.Close()

	handler := makeQueryTimeSeriesHandler(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"metric":    "bytes",
		"dimension": "AS_dst",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success result, got error: %v", result.Content)
	}
}
