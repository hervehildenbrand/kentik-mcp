package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/hervehildenbrand/kentik-mcp/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
)

func testClientWithServer(url string) *client.Client {
	return client.New(&config.Config{
		Email:    "test@example.com",
		APIToken: "test-token",
		V5Base:   url,
		V6Base:   url,
	})
}

func TestListDevicesHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/devices" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Write([]byte(`{"devices":[{"id":"1","device_name":"router1"}]}`))
	}))
	defer srv.Close()

	handler := makeListDevicesHandler(testClientWithServer(srv.URL))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "router1") {
		t.Errorf("expected response to contain router1, got: %s", text)
	}
}

func TestGetDeviceHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/device/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"device":{"id":"42","device_name":"core-rtr"}}`))
	}))
	defer srv.Close()

	handler := makeGetDeviceHandler(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"device_id": "42"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "core-rtr") {
		t.Errorf("expected response to contain core-rtr, got: %s", text)
	}
}

func TestGetDeviceHandler_MissingParam(t *testing.T) {
	handler := makeGetDeviceHandler(testClientWithServer("http://unused"))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result when device_id missing")
	}
}
