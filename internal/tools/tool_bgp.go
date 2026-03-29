package tools

import (
	"context"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerBGPTool(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("kentik_bgp",
		mcp.WithDescription("Query Kentik BGP routing data. Actions: "+
			"'list_monitors' — BGP prefix monitors tracking routing changes and hijacks. "+
			"'get_routes' — BGP route observations from global vantage points (requires monitor_id, optional prefix)."),
		mcp.WithString("action", mcp.Required(),
			mcp.Description("One of: list_monitors, get_routes")),
		mcp.WithString("monitor_id", mcp.Description("Monitor ID for get_routes")),
		mcp.WithString("prefix", mcp.Description("IP prefix for get_routes (e.g. '8.8.8.0/24')")),
	), makeBGPRouter(c))
}

func makeBGPRouter(c *client.Client) server.ToolHandlerFunc {
	listMonitors := makeListBGPMonitorsHandler(c)
	getRoutes := makeGetBGPRoutesHandler(c)

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		action, err := request.RequireString("action")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		switch action {
		case "list_monitors":
			return listMonitors(ctx, request)
		case "get_routes":
			return getRoutes(ctx, request)
		default:
			return mcp.NewToolResultError("Unknown action: " + action), nil
		}
	}
}
