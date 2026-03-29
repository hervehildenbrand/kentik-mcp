package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestSyntheticsRouter_ListTests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/synthetics/v202309/tests" {
			w.Write([]byte(`{"tests":[{"id":"1","name":"DNS Test","type":"dns_grid","status":"active"}]}`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeSyntheticsRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"action": "list_tests"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "DNS Test") {
		t.Errorf("expected test name in response, got: %s", text)
	}
}

func TestSyntheticsRouter_GetTest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/synthetics/v202309/tests/42" {
			w.Write([]byte(`{"test":{"id":"42","name":"HTTP Check","type":"http"}}`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeSyntheticsRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"action":  "get_test",
		"test_id": "42",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
}

func TestSyntheticsRouter_GetResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/synthetics/v202309/results" && r.Method == "POST" {
			w.Write([]byte(`{"results":[{"testId":"1","health":"healthy","agents":[]}]}`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeSyntheticsRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"action":     "get_results",
		"test_ids":   "1,2",
		"start_time": "2026-03-28T12:00:00Z",
		"end_time":   "2026-03-28T13:00:00Z",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
}

func TestSyntheticsRouter_ListAgents(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/synthetics/v202309/agents" {
			w.Write([]byte(`{"agents":[{"id":"1","city":"Seattle","country":"US","type":"global","status":"online","ip":"1.2.3.4"}]}`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeSyntheticsRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"action": "list_agents"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "Seattle") {
		t.Errorf("expected agent city in response, got: %s", text)
	}
}

func TestSyntheticsRouter_ListAgents_FilterByType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/synthetics/v202309/agents" {
			w.Write([]byte(`{"agents":[{"id":"1","city":"Seattle","type":"global"},{"id":"2","city":"Private DC","type":"private"}]}`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeSyntheticsRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"action": "list_agents",
		"type":   "global",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "Seattle") {
		t.Error("expected Seattle global agent")
	}
	if strings.Contains(text, "Private DC") {
		t.Error("private agent should be filtered out")
	}
}

func TestSyntheticsRouter_UnknownAction(t *testing.T) {
	handler := makeSyntheticsRouter(testClientWithServer("http://unused"))
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
