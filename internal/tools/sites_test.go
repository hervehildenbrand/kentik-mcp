package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestListSitesHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sites" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Write([]byte(`{"sites":[{"id":"1","site_name":"DC-East","lat":40.7128,"lon":-74.006}]}`))
	}))
	defer srv.Close()

	handler := makeListSitesHandler(testClientWithServer(srv.URL))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "DC-East") {
		t.Errorf("expected response to contain DC-East, got: %s", text)
	}
}

func TestGetSiteHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/site/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"site":{"id":"42","site_name":"DC-West","lat":37.7749,"lon":-122.4194}}`))
	}))
	defer srv.Close()

	handler := makeGetSiteHandler(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"site_id": "42"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "DC-West") {
		t.Errorf("expected response to contain DC-West, got: %s", text)
	}
}

func TestGetSiteHandler_MissingParam(t *testing.T) {
	handler := makeGetSiteHandler(testClientWithServer("http://unused"))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result when site_id missing")
	}
}
