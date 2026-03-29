package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestListNotificationChannelsHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/notification_channel/v202204beta1/channels" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Write([]byte(`{"channels":[{"id":"1","name":"ops-slack","type":"slack"}]}`))
	}))
	defer srv.Close()

	handler := makeListNotificationChannelsHandler(testClientWithServer(srv.URL))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "ops-slack") {
		t.Errorf("expected response to contain ops-slack, got: %s", text)
	}
}

func TestGetNotificationChannelHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/notification_channel/v202204beta1/channels/99" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"channel":{"id":"99","name":"pagerduty-critical","type":"pagerduty","pagerdutyConfig":{"serviceKey":"xxx"}}}`))
	}))
	defer srv.Close()

	handler := makeGetNotificationChannelHandler(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"channel_id": "99"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "pagerduty-critical") {
		t.Errorf("expected response to contain pagerduty-critical, got: %s", text)
	}
}

func TestGetNotificationChannelHandler_MissingParam(t *testing.T) {
	handler := makeGetNotificationChannelHandler(testClientWithServer("http://unused"))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result when channel_id missing")
	}
}
