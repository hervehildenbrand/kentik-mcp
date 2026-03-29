package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestListAlertPoliciesHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/alerting/policies" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"policies":[{"id":"123","name":"DDoS SYN Flood","status":"active","dimensions":["IP_dst"],"metrics":["packets","bits"]}],"count":1}`))
	}))
	defer srv.Close()

	handler := makeListAlertPoliciesHandler(testClientWithServer(srv.URL))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "DDoS SYN Flood") {
		t.Errorf("expected policy name in response, got: %s", text)
	}
}

func TestListAlertPoliciesHandler_FilterByStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"policies":[{"id":"1","name":"Active Policy","status":"active","dimensions":[],"metrics":[]},{"id":"2","name":"Disabled Policy","status":"disabled","dimensions":[],"metrics":[]}],"count":2}`))
	}))
	defer srv.Close()

	handler := makeListAlertPoliciesHandler(testClientWithServer(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"status": "active"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "Active Policy") {
		t.Error("expected active policy in filtered results")
	}
	if strings.Contains(text, "Disabled Policy") {
		t.Error("disabled policy should be filtered out")
	}
}

func TestListActiveAlarmsHandler_NoAlarms(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/alerts-active/alarms") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	handler := makeListActiveAlarmsHandler(testClientWithServer(srv.URL))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "No active alarms") {
		t.Errorf("expected no alarms message, got: %s", text)
	}
}

func TestListAlertHistoryHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/alerts-active/alerts-history") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`[{"alarm_start":"2026-03-25T13:14:07Z","policy_name":"DDoS Volumetric Protection","alert_key":"198.51.100.0/24","alert_severity":"minor2","alarm_state":"CLEAR","alert_value":12575166570.0,"alert_metric":["bits","packets"]}]`))
	}))
	defer srv.Close()

	handler := makeListAlertHistoryHandler(testClientWithServer(srv.URL))
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "DDoS Volumetric") {
		t.Errorf("expected policy name in response, got: %s", text)
	}
	if !strings.Contains(text, "198.51.100.0/24") {
		t.Errorf("expected target prefix in response, got: %s", text)
	}
}
