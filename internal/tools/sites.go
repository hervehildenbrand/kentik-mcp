package tools

import (
	"context"
	"fmt"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerSiteTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("list_sites",
		mcp.WithDescription("List all sites (physical locations like data centers and PoPs) in your Kentik organization. Returns site name, coordinates (lat/lon), and ID. Sites group devices by physical location — use site names as dimensions in query_flows with dimension 'i_device_site_name'."),
	), makeListSitesHandler(c))

	s.AddTool(mcp.NewTool("get_site",
		mcp.WithDescription("Get detailed information about a specific site including name, latitude, longitude, and company ID."),
		mcp.WithString("site_id", mcp.Required(), mcp.Description("The site ID")),
	), makeGetSiteHandler(c))
}

func makeListSitesHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := c.V5("GET", "/sites", nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list sites: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}

func makeGetSiteHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		siteID, err := request.RequireString("site_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := c.V5("GET", fmt.Sprintf("/site/%s", siteID), nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get site: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}
