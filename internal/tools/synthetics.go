package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerSyntheticsTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("list_synthetic_tests",
		mcp.WithDescription("List all synthetic monitoring tests. Test types: dns_grid (DNS resolution from multiple servers), http (URL monitoring), page_load (full browser), network_mesh (agent-to-agent latency), hostname (ping). Shows test name, type, status (active/paused), and target. Use get_synthetic_results with a test ID and time range for actual measurements."),
	), makeListSyntheticTestsHandler(c))

	s.AddTool(mcp.NewTool("get_synthetic_test",
		mcp.WithDescription("Get full configuration of a synthetic test: target host/URL, DNS servers, assigned agents (by ID — cross-reference with list_synthetic_agents), health thresholds (latency warning/critical in microseconds), notification channels, and test frequency."),
		mcp.WithString("test_id", mcp.Required(),
			mcp.Description("The synthetic test ID (from list_synthetic_tests)")),
	), makeGetSyntheticTestHandler(c))

	s.AddTool(mcp.NewTool("get_synthetic_results",
		mcp.WithDescription("Get synthetic test results over a time range. Returns per-agent, per-task measurements: DNS resolution latency and response data, HTTP status codes and latency, ping RTT. Each entry includes health status (healthy/warning/critical) based on configured thresholds. Results are grouped by time bucket."),
		mcp.WithString("test_ids", mcp.Required(),
			mcp.Description("Comma-separated test IDs (from list_synthetic_tests)")),
		mcp.WithString("start_time", mcp.Required(),
			mcp.Description("Start time in ISO format UTC (e.g. '2026-03-28T12:00:00Z')")),
		mcp.WithString("end_time", mcp.Required(),
			mcp.Description("End time in ISO format (e.g. '2026-03-28T13:00:00Z')")),
	), makeGetSyntheticResultsHandler(c))

	s.AddTool(mcp.NewTool("list_synthetic_agents",
		mcp.WithDescription("List synthetic monitoring agents available for running tests. Kentik provides ~270 global agents in cities worldwide (Madrid, Seattle, Paris, Tokyo, etc.). Private agents are customer-deployed. Each agent has an ID, city, country, IP, and status. Agent IDs are referenced in synthetic test configurations."),
		mcp.WithString("type",
			mcp.Description("Filter by agent type: 'global' (Kentik-hosted worldwide) or 'private' (customer-deployed). Empty for all.")),
	), makeListSyntheticAgentsHandler(c))
}

func makeListSyntheticTestsHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := c.V6("GET", "/synthetics/v202309/tests", nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list synthetic tests: %v", err)), nil
		}

		var resp struct {
			Tests []map[string]any `json:"tests"`
		}
		if err := json.Unmarshal(data, &resp); err != nil {
			return mcp.NewToolResultText(formatJSON(data)), nil
		}

		if len(resp.Tests) == 0 {
			return mcp.NewToolResultText("No synthetic tests configured."), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("## Synthetic Tests (%d)\n\n", len(resp.Tests)))
		sb.WriteString("| ID | Name | Type | Status | Target |\n")
		sb.WriteString("|----|------|------|--------|--------|\n")

		for _, t := range resp.Tests {
			target := extractTarget(t)
			sb.WriteString(fmt.Sprintf("| %v | %v | %v | %v | %s |\n",
				t["id"], t["name"], t["type"], t["status"], target))
		}

		return mcp.NewToolResultText(sb.String()), nil
	}
}

func makeGetSyntheticTestHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		testID, err := request.RequireString("test_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, err := c.V6("GET", fmt.Sprintf("/synthetics/v202309/tests/%s", testID), nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get test: %v", err)), nil
		}

		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}

func makeGetSyntheticResultsHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		testIdsStr, err := request.RequireString("test_ids")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		startTime, err := request.RequireString("start_time")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		endTime, err := request.RequireString("end_time")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		ids := parseDimensions(testIdsStr)
		body := map[string]any{
			"ids":       ids,
			"startTime": startTime,
			"endTime":   endTime,
			"augmented": true,
		}

		data, err := c.V6("POST", "/synthetics/v202309/results", body)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get results: %v", err)), nil
		}

		var resp struct {
			Results []map[string]any `json:"results"`
		}
		if err := json.Unmarshal(data, &resp); err != nil {
			return mcp.NewToolResultText(formatJSON(data)), nil
		}

		if len(resp.Results) == 0 {
			return mcp.NewToolResultText("No results found for the specified time range."), nil
		}

		return mcp.NewToolResultText(summarizeSyntheticResults(resp.Results)), nil
	}
}

func makeListSyntheticAgentsHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := c.V6("GET", "/synthetics/v202309/agents", nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list agents: %v", err)), nil
		}

		var resp struct {
			Agents []map[string]any `json:"agents"`
		}
		if err := json.Unmarshal(data, &resp); err != nil {
			return mcp.NewToolResultText(formatJSON(data)), nil
		}

		typeFilter := request.GetString("type", "")
		agents := resp.Agents
		if typeFilter != "" {
			var filtered []map[string]any
			for _, a := range agents {
				if fmt.Sprintf("%v", a["type"]) == typeFilter {
					filtered = append(filtered, a)
				}
			}
			agents = filtered
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("## Synthetic Agents (%d)\n\n", len(agents)))
		sb.WriteString("| ID | City | Country | Type | Status | IP |\n")
		sb.WriteString("|----|------|---------|------|--------|----|\n")

		for _, a := range agents {
			sb.WriteString(fmt.Sprintf("| %v | %v | %v | %v | %v | %v |\n",
				a["id"], a["city"], a["country"], a["type"], a["status"], a["ip"]))
		}

		return mcp.NewToolResultText(sb.String()), nil
	}
}

func extractTarget(test map[string]any) string {
	settings, ok := test["settings"].(map[string]any)
	if !ok {
		return ""
	}
	// Check various test type settings
	for _, key := range []string{"dnsGrid", "dns", "http", "pageLoad", "networkMesh", "hostname"} {
		if cfg, ok := settings[key].(map[string]any); ok {
			if target, ok := cfg["target"].(string); ok {
				return target
			}
		}
	}
	return ""
}

func summarizeSyntheticResults(results []map[string]any) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Synthetic Results (%d entries)\n\n", len(results)))

	for _, r := range results {
		testID := r["testId"]
		health := r["health"]
		sb.WriteString(fmt.Sprintf("### Test %v — Health: %v\n\n", testID, health))

		agents, ok := r["agents"].([]any)
		if !ok {
			continue
		}

		sb.WriteString("| Agent | Health | Task | Latency | Status |\n")
		sb.WriteString("|-------|--------|------|---------|--------|\n")

		for _, agent := range agents {
			a, ok := agent.(map[string]any)
			if !ok {
				continue
			}
			agentID := a["agentId"]
			agentHealth := a["health"]

			tasks, ok := a["tasks"].([]any)
			if !ok {
				continue
			}
			for _, task := range tasks {
				t, ok := task.(map[string]any)
				if !ok {
					continue
				}
				taskHealth := t["health"]

				// Try DNS task
				if dns, ok := t["dns"].(map[string]any); ok {
					latency := "?"
					if lat, ok := dns["latency"].(map[string]any); ok {
						if cur, ok := lat["current"].(float64); ok {
							latency = fmt.Sprintf("%.1fms", cur/1000)
						}
					}
					server := dns["server"]
					sb.WriteString(fmt.Sprintf("| %v | %v | DNS→%v | %s | %v |\n",
						agentID, agentHealth, server, latency, taskHealth))
				}

				// Try HTTP task
				if http, ok := t["http"].(map[string]any); ok {
					latency := "?"
					if lat, ok := http["latency"].(map[string]any); ok {
						if cur, ok := lat["current"].(float64); ok {
							latency = fmt.Sprintf("%.1fms", cur/1000)
						}
					}
					sb.WriteString(fmt.Sprintf("| %v | %v | HTTP | %s | %v |\n",
						agentID, agentHealth, latency, taskHealth))
				}

				// Try ping task
				if ping, ok := t["ping"].(map[string]any); ok {
					latency := "?"
					if lat, ok := ping["latency"].(map[string]any); ok {
						if cur, ok := lat["current"].(float64); ok {
							latency = fmt.Sprintf("%.1fms", cur/1000)
						}
					}
					sb.WriteString(fmt.Sprintf("| %v | %v | Ping | %s | %v |\n",
						agentID, agentHealth, latency, taskHealth))
				}
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
