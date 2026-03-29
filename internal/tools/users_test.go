package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestListUsersHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Write([]byte(`{"users":[{"id":"1","user_email":"admin@example.com","user_full_name":"Admin User","role":"Administrator"}]}`))
	}))
	defer srv.Close()

	handler := makeListUsersHandler(testClientWithServer(srv.URL))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "admin@example.com") {
		t.Errorf("expected response to contain admin@example.com, got: %s", text)
	}
}

func TestGetUserHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"user":{"id":"42","user_email":"user@example.com","user_full_name":"Test User","role":"Member"}}`))
	}))
	defer srv.Close()

	handler := makeGetUserHandler(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"user_id": "42"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "user@example.com") {
		t.Errorf("expected response to contain user@example.com, got: %s", text)
	}
}

func TestGetUserHandler_MissingParam(t *testing.T) {
	handler := makeGetUserHandler(testClientWithServer("http://unused"))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result when user_id missing")
	}
}
