package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestListLabelsHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/label/v202210/labels" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Write([]byte(`{"labels":[{"id":"1","name":"inet","color":"#FF0000"}]}`))
	}))
	defer srv.Close()

	handler := makeListLabelsHandler(testClientWithServer(srv.URL))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "inet") {
		t.Errorf("expected response to contain inet, got: %s", text)
	}
}

func TestGetLabelHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/label/v202210/labels/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"label":{"id":"42","name":"emea","color":"#00FF00"}}`))
	}))
	defer srv.Close()

	handler := makeGetLabelHandler(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"label_id": "42"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "emea") {
		t.Errorf("expected response to contain emea, got: %s", text)
	}
}

func TestGetLabelHandler_MissingParam(t *testing.T) {
	handler := makeGetLabelHandler(testClientWithServer("http://unused"))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result when label_id missing")
	}
}
