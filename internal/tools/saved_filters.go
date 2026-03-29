package tools

import (
	"context"
	"fmt"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerSavedFilterTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("list_saved_filters",
		mcp.WithDescription("List saved filter definitions that can be reused across flow queries. Each saved filter contains filter groups with conditions (field, operator, value). Use list_saved_filters to discover available filters, then use their filter structure in query_flows 'filters_json' parameter for consistent query filtering."),
	), makeListSavedFiltersHandler(c))

	s.AddTool(mcp.NewTool("get_saved_filter",
		mcp.WithDescription("Get the full definition of a saved filter including all filter groups, conditions, operators, and values."),
		mcp.WithString("filter_id", mcp.Required(), mcp.Description("The saved filter ID")),
	), makeGetSavedFilterHandler(c))
}

func makeListSavedFiltersHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := c.V5("GET", "/saved-filters/custom", nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list saved filters: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}

func makeGetSavedFilterHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filterID, err := request.RequireString("filter_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := c.V5("GET", fmt.Sprintf("/saved-filter/custom/%s", filterID), nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get saved filter: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}
