package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestAlertsRouter_ListPolicies(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/alerting/policies" {
			w.Write([]byte(`{"policies":[{"id":"1","name":"DDoS SYN Flood","status":"active","dimensions":["IP_dst"],"metrics":["packets"]}],"count":1}`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeAlertsRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"action": "list_policies"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "DDoS SYN Flood") {
		t.Errorf("expected policy name in response, got: %s", text)
	}
}

func TestAlertsRouter_GetPolicy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/alerting/policies" {
			w.Write([]byte(`{"policies":[{"id":"123","name":"Shield Policy","status":"active","dimensions":["IP_dst"],"metrics":["bits"]}],"count":1}`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeAlertsRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"action":    "get_policy",
		"policy_id": "123",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "Shield Policy") {
		t.Errorf("expected policy name in response, got: %s", text)
	}
}

func TestAlertsRouter_ActiveAlarms(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/alerts-active/alarms") {
			w.Write([]byte(`[]`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeAlertsRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"action": "active_alarms"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "No active alarms") {
		t.Errorf("expected no alarms message, got: %s", text)
	}
}

func TestAlertsRouter_AlertHistory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/alerts-active/alerts-history") {
			w.Write([]byte(`[{"alarm_start":"2026-03-25T13:14:07Z","policy_name":"Shield","alert_key":"1.2.3.4","alert_severity":"minor","alarm_state":"CLEAR","alert_value":1000000}]`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeAlertsRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"action": "alert_history"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
}

func TestAlertsRouter_TopAttacks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/alerts-active/alerts-history") {
			w.Write([]byte(`[{"alarm_start":"2026-03-25T13:14:07Z","alarm_end":"2026-03-25T13:24:07Z","policy_name":"DDoS","alert_key":"target","alert_severity":"critical","alert_value":5000000000,"alert_metric":["bits"]}]`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	handler := makeAlertsRouter(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"action":  "top_attacks",
		"sort_by": "bps",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
}

func TestAlertsRouter_UnknownAction(t *testing.T) {
	handler := makeAlertsRouter(testClientWithServer("http://unused"))
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
