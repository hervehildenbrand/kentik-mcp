package tools

import (
	"context"
	"fmt"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerKMITools(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("get_as_info",
		mcp.WithDescription("Look up any Autonomous System (AS) number to get its name, country, organization, and network relationship summary. Useful for understanding who is behind traffic seen in flow queries. Note: KMI API may not be available on all Kentik regions."),
		mcp.WithString("asn", mcp.Required(),
			mcp.Description("The AS number like '15169'")),
	), makeGetASInfoHandler(c))

	s.AddTool(mcp.NewTool("get_as_relationships",
		mcp.WithDescription("Get the transit providers, customers, or settlement-free peers of an AS number. Useful for understanding routing paths, peering decisions, and upstream/downstream dependencies. Returns ranked list of related ASNs."),
		mcp.WithString("asn", mcp.Required(),
			mcp.Description("The AS number like '15169'")),
		mcp.WithString("relationship_type", mcp.Required(),
			mcp.Description("One of: 'customers', 'providers', 'peers'")),
	), makeGetASRelationshipsHandler(c))
}

func makeGetASInfoHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		asn, err := request.RequireString("asn")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := c.V6("GET", fmt.Sprintf("/kmi/v202212/as/%s", asn), nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get AS info: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}

func makeGetASRelationshipsHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		asn, err := request.RequireString("asn")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		relType, err := request.RequireString("relationship_type")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := c.V6("GET", fmt.Sprintf("/kmi/v202212/as/%s/%s", asn, relType), nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get AS relationships: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}
