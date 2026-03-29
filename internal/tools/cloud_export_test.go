package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestListCloudExportsHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud_export/v202101beta1/exports" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Write([]byte(`{"exports":[{"id":"1","name":"aws-vpc-prod","cloudProvider":"aws","status":"ACTIVE"}]}`))
	}))
	defer srv.Close()

	handler := makeListCloudExportsHandler(testClientWithServer(srv.URL))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "aws-vpc-prod") {
		t.Errorf("expected response to contain aws-vpc-prod, got: %s", text)
	}
}

func TestGetCloudExportHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud_export/v202101beta1/exports/123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"export":{"id":"123","name":"gcp-flow-logs","cloudProvider":"gcp","gcpProperties":{"project":"my-project"}}}`))
	}))
	defer srv.Close()

	handler := makeGetCloudExportHandler(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"export_id": "123"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "gcp-flow-logs") {
		t.Errorf("expected response to contain gcp-flow-logs, got: %s", text)
	}
}

func TestGetCloudExportHandler_MissingParam(t *testing.T) {
	handler := makeGetCloudExportHandler(testClientWithServer("http://unused"))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result when export_id missing")
	}
}
