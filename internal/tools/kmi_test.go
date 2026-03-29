package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestGetASInfoHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/kmi/v202212/as/15169" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Write([]byte(`{"asn":15169,"name":"GOOGLE","country":"US","orgName":"Google LLC"}`))
	}))
	defer srv.Close()

	handler := makeGetASInfoHandler(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"asn": "15169"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "GOOGLE") {
		t.Errorf("expected response to contain GOOGLE, got: %s", text)
	}
}

func TestGetASInfoHandler_MissingParam(t *testing.T) {
	handler := makeGetASInfoHandler(testClientWithServer("http://unused"))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result when asn missing")
	}
}

func TestGetASRelationshipsHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/kmi/v202212/as/15169/providers" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"asn":15169,"relationshipType":"providers","relationships":[{"asn":174,"name":"COGENT-174"}]}`))
	}))
	defer srv.Close()

	handler := makeGetASRelationshipsHandler(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"asn": "15169", "relationship_type": "providers"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "COGENT-174") {
		t.Errorf("expected response to contain COGENT-174, got: %s", text)
	}
}

func TestGetASRelationshipsHandler_MissingASN(t *testing.T) {
	handler := makeGetASRelationshipsHandler(testClientWithServer("http://unused"))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"relationship_type": "providers"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result when asn missing")
	}
}

func TestGetASRelationshipsHandler_MissingRelType(t *testing.T) {
	handler := makeGetASRelationshipsHandler(testClientWithServer("http://unused"))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"asn": "15169"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result when relationship_type missing")
	}
}
