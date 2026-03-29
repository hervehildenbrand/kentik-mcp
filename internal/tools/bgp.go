package tools

import (
	"context"
	"fmt"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerBGPTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("list_bgp_monitors",
		mcp.WithDescription("List all BGP monitors configured in Kentik. BGP monitors track routing changes and anomalies for specified IP prefixes from global vantage points. Each monitor watches up to 10 prefixes and reports reachability, AS path changes, and route hijacking. Use get_bgp_routes with a monitor ID to see current routing state."),
	), makeListBGPMonitorsHandler(c))

	s.AddTool(mcp.NewTool("get_bgp_routes",
		mcp.WithDescription("Get BGP route observations for a monitored prefix. Returns AS paths, origin AS, and reachability data from Kentik's BGP vantage points worldwide. Useful for verifying routing, detecting hijacks, and understanding path diversity."),
		mcp.WithString("monitor_id", mcp.Required(),
			mcp.Description("The BGP monitor ID (from list_bgp_monitors)")),
		mcp.WithString("prefix",
			mcp.Description("IP prefix to look up routes for (e.g. '8.8.8.0/24')")),
	), makeGetBGPRoutesHandler(c))
}

func makeListBGPMonitorsHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := c.V6("GET", "/bgp_monitoring/v202205beta1/monitors", nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list BGP monitors: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}

func makeGetBGPRoutesHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		monitorID, err := request.RequireString("monitor_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{}
		if prefix := request.GetString("prefix", ""); prefix != "" {
			body["prefix"] = prefix
		}

		data, err := c.V6("POST", fmt.Sprintf("/bgp_monitoring/v202205beta1/monitors/%s/routes", monitorID), body)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get BGP routes: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}
