package tools

import (
	"context"
	"fmt"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerCustomDimensionTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("list_custom_dimensions",
		mcp.WithDescription("List custom dimensions that enrich flow data with business context (e.g., mapping IPs to application names, customer IDs, or cost centers). Custom dimensions have 'populators' — rules that match traffic and assign values. Custom dimension names can be used as group-by dimensions in query_flows."),
	), makeListCustomDimensionsHandler(c))

	s.AddTool(mcp.NewTool("get_custom_dimension",
		mcp.WithDescription("Get a custom dimension's configuration including its populator rules that define how traffic is classified."),
		mcp.WithString("dimension_id", mcp.Required(),
			mcp.Description("The custom dimension ID")),
	), makeGetCustomDimensionHandler(c))
}

func makeListCustomDimensionsHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := c.V5("GET", "/customdimensions", nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list custom dimensions: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}

func makeGetCustomDimensionHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		dimensionID, err := request.RequireString("dimension_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := c.V5("GET", fmt.Sprintf("/customdimension/%s", dimensionID), nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get custom dimension: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}
