package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestMarketRouter_ASInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/kmi/v202212/as/15169" {
			w.Write([]byte(`{"asn":"15169","name":"Google LLC","country":"US"}`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeMarketRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"action": "as_info",
		"asn":    "15169",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "Google") {
		t.Errorf("expected Google in response, got: %s", text)
	}
}

func TestMarketRouter_ASRelationships(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/kmi/v202212/as/15169/customers" {
			w.Write([]byte(`{"asn":"15169","customers":[{"asn":"12345","name":"Customer AS"}]}`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeMarketRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"action":            "as_relationships",
		"asn":               "15169",
		"relationship_type": "customers",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
}

func TestMarketRouter_ASRelationships_Providers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/kmi/v202212/as/64500/providers" {
			w.Write([]byte(`{"asn":"64500","providers":[{"asn":"3356","name":"Lumen"}]}`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeMarketRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"action":            "as_relationships",
		"asn":               "64500",
		"relationship_type": "providers",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
}

func TestMarketRouter_UnknownAction(t *testing.T) {
	handler := makeMarketRouter(testClientWithServer("http://unused"))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"action": "invalid",
		"asn":    "15169",
	}

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

func TestMarketRouter_MissingASN(t *testing.T) {
	handler := makeMarketRouter(testClientWithServer("http://unused"))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"action": "as_info",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error when ASN is missing")
	}
}
