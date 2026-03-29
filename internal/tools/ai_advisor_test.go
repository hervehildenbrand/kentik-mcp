package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestAIAdvisorHandler(t *testing.T) {
	var pollCount int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && strings.Contains(r.URL.Path, "/chat") {
			w.Write([]byte(`{"id":"test-session-123","status":"SESSION_STATUS_PENDING"}`))
			return
		}
		if r.Method == "GET" && strings.Contains(r.URL.Path, "test-session-123") {
			count := atomic.AddInt32(&pollCount, 1)
			if count < 2 {
				w.Write([]byte(`{"id":"test-session-123","status":"SESSION_STATUS_PROCESSING","messages":[{"status":"SESSION_STATUS_PROCESSING","finalAnswer":"","reasoning":"thinking..."}]}`))
			} else {
				w.Write([]byte(`{"id":"test-session-123","status":"SESSION_STATUS_COMPLETE","messages":[{"status":"SESSION_STATUS_COMPLETE","finalAnswer":"Your top source is AS15169 at 3.2 Gbps.","reasoning":"done"}]}`))
			}
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	handler := makeAIAdvisorHandler(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"question": "What are my top sources?"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "AS15169") {
		t.Errorf("expected AI response in result, got: %s", text)
	}
}
