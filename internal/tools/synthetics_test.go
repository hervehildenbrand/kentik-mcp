package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestListSyntheticTestsHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/synthetics/v202309/tests" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"tests":[{"id":"123","name":"DNS Test","type":"dns_grid","status":"TEST_STATUS_ACTIVE","settings":{"dnsGrid":{"target":"example.com"}}}]}`))
	}))
	defer srv.Close()

	handler := makeListSyntheticTestsHandler(testClientWithServer(srv.URL))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "DNS Test") {
		t.Errorf("expected test name in response, got: %s", text)
	}
	if !strings.Contains(text, "example.com") {
		t.Errorf("expected target in response, got: %s", text)
	}
}

func TestListSyntheticAgentsHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"agents":[{"id":"1","city":"Paris","country":"FR","type":"global","status":"AGENT_STATUS_OK","ip":"1.2.3.4"}]}`))
	}))
	defer srv.Close()

	handler := makeListSyntheticAgentsHandler(testClientWithServer(srv.URL))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "Paris") {
		t.Errorf("expected agent city in response, got: %s", text)
	}
}

func TestGetSyntheticResultsHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"results":[{"testId":"123","health":"healthy","agents":[{"agentId":"1","health":"healthy","tasks":[{"dns":{"target":"example.com","server":"8.8.8.8","latency":{"current":5000}},"health":"healthy"}]}]}]}`))
	}))
	defer srv.Close()

	handler := makeGetSyntheticResultsHandler(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"test_ids":   "123",
		"start_time": "2026-03-28T12:00:00Z",
		"end_time":   "2026-03-28T13:00:00Z",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "healthy") {
		t.Errorf("expected health status in response, got: %s", text)
	}
}
