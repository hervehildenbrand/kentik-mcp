package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestAITool_Handler(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ai_advisor/v202511/chat" && r.Method == "POST" {
			w.Write([]byte(`{"id":"session123","status":"SESSION_STATUS_PENDING"}`))
		} else if strings.HasPrefix(r.URL.Path, "/ai_advisor/v202511/chat/session123") {
			callCount++
			if callCount >= 2 {
				w.Write([]byte(`{"id":"session123","status":"SESSION_STATUS_COMPLETED","messages":[{"status":"completed","finalAnswer":"Your traffic is healthy."}]}`))
			} else {
				w.Write([]byte(`{"id":"session123","status":"SESSION_STATUS_PROCESSING"}`))
			}
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeAIAdvisorHandler(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"question": "What is my traffic like?",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result, err := handler(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "traffic is healthy") {
		t.Errorf("expected AI answer in response, got: %s", text)
	}
}

func TestAITool_MissingQuestion(t *testing.T) {
	handler := makeAIAdvisorHandler(testClientWithServer("http://unused"))
	req := mcp.CallToolRequest{}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error when question is missing")
	}
}
