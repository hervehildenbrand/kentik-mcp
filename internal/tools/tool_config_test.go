package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestConfigRouter_ListSavedFilters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/saved-filters/custom" {
			w.Write([]byte(`[{"id":"1","filter_name":"Production Traffic"}]`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeConfigRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"action": "list_saved_filters"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
}

func TestConfigRouter_GetSavedFilter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/saved-filter/custom/42" {
			w.Write([]byte(`{"id":"42","filter_name":"Web Traffic","filterGroups":[]}`))
		} else {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeConfigRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"action": "get_saved_filter",
		"id":     "42",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
}

func TestConfigRouter_ListCustomDimensions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/customdimensions" {
			w.Write([]byte(`{"customDimensions":[{"id":"1","name":"App Name"}]}`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeConfigRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"action": "list_custom_dimensions"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
}

func TestConfigRouter_GetCustomDimension(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/customdimension/5" {
			w.Write([]byte(`{"customDimension":{"id":"5","name":"Customer ID"}}`))
		} else {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeConfigRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"action": "get_custom_dimension",
		"id":     "5",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
}

func TestConfigRouter_ListUsers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users" {
			w.Write([]byte(`{"users":[{"id":"1","user_email":"admin@example.com"}]}`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeConfigRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"action": "list_users"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
}

func TestConfigRouter_GetUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user/99" {
			w.Write([]byte(`{"user":{"id":"99","user_email":"ops@example.com"}}`))
		} else {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeConfigRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"action": "get_user",
		"id":     "99",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
}

func TestConfigRouter_ListCloudExports(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cloud_export/v202101beta1/exports" {
			w.Write([]byte(`{"exports":[{"id":"1","name":"AWS VPC Flow Logs"}]}`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeConfigRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"action": "list_cloud_exports"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
}

func TestConfigRouter_GetCloudExport(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cloud_export/v202101beta1/exports/7" {
			w.Write([]byte(`{"export":{"id":"7","name":"Azure Flow Logs"}}`))
		} else {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeConfigRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"action": "get_cloud_export",
		"id":     "7",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
}

func TestConfigRouter_ListNotificationChannels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/notification_channel/v202204beta1/channels" {
			w.Write([]byte(`{"channels":[{"id":"1","name":"Slack Alerts"}]}`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeConfigRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"action": "list_notification_channels"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
}

func TestConfigRouter_GetNotificationChannel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/notification_channel/v202204beta1/channels/3" {
			w.Write([]byte(`{"channel":{"id":"3","name":"PagerDuty"}}`))
		} else {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeConfigRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"action": "get_notification_channel",
		"id":     "3",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
}

func TestConfigRouter_UnknownAction(t *testing.T) {
	handler := makeConfigRouter(testClientWithServer("http://unused"))
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

func TestRemapID(t *testing.T) {
	called := false
	var receivedKey string

	mockHandler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		called = true
		receivedKey = request.GetString("filter_id", "")
		return mcp.NewToolResultText("ok"), nil
	}

	wrapped := remapID(mockHandler, "filter_id")
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"id": "42"}

	_, err := wrapped(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
	if receivedKey != "42" {
		t.Errorf("expected filter_id=42, got: %s", receivedKey)
	}
}
