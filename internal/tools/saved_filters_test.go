package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestListSavedFiltersHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/saved-filters/custom" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Write([]byte(`{"filters":[{"id":"1","filter_name":"Internal Traffic"}]}`))
	}))
	defer srv.Close()

	handler := makeListSavedFiltersHandler(testClientWithServer(srv.URL))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "Internal Traffic") {
		t.Errorf("expected response to contain Internal Traffic, got: %s", text)
	}
}

func TestGetSavedFilterHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/saved-filter/custom/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"filter":{"id":"42","filter_name":"External Traffic","filters":{"filterGroups":[{"filters":[{"filterField":"src_as","operator":"=","filterValue":"15169"}]}]}}}`))
	}))
	defer srv.Close()

	handler := makeGetSavedFilterHandler(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"filter_id": "42"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "External Traffic") {
		t.Errorf("expected response to contain External Traffic, got: %s", text)
	}
}

func TestGetSavedFilterHandler_MissingParam(t *testing.T) {
	handler := makeGetSavedFilterHandler(testClientWithServer("http://unused"))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result when filter_id missing")
	}
}
