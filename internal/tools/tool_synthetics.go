package tools

import (
	"context"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerSyntheticsTool(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("kentik_synthetics",
		mcp.WithDescription("Query Kentik synthetic monitoring. Actions: "+
			"'list_tests' — all tests (DNS, HTTP, ping, mesh) with status and target. "+
			"'get_test' — test config detail (requires test_id). "+
			"'get_results' — per-agent results over time range (requires test_ids + start_time + end_time). "+
			"'list_agents' — available test agents (~270 global, filter by type)."),
		mcp.WithString("action", mcp.Required(),
			mcp.Description("One of: list_tests, get_test, get_results, list_agents")),
		mcp.WithString("test_id", mcp.Description("Test ID for get_test")),
		mcp.WithString("test_ids", mcp.Description("Comma-separated test IDs for get_results")),
		mcp.WithString("start_time", mcp.Description("ISO start time for get_results (e.g. '2026-03-28T12:00:00Z')")),
		mcp.WithString("end_time", mcp.Description("ISO end time for get_results")),
		mcp.WithString("type", mcp.Description("Agent filter for list_agents: 'global' or 'private'")),
	), makeSyntheticsRouter(c))
}

func makeSyntheticsRouter(c *client.Client) server.ToolHandlerFunc {
	listTests := makeListSyntheticTestsHandler(c)
	getTest := makeGetSyntheticTestHandler(c)
	getResults := makeGetSyntheticResultsHandler(c)
	listAgents := makeListSyntheticAgentsHandler(c)

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		action, err := request.RequireString("action")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		switch action {
		case "list_tests":
			return listTests(ctx, request)
		case "get_test":
			return getTest(ctx, request)
		case "get_results":
			return getResults(ctx, request)
		case "list_agents":
			return listAgents(ctx, request)
		default:
			return mcp.NewToolResultError("Unknown action: " + action), nil
		}
	}
}
