package tools

import (
	"context"
	"fmt"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerLabelTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("list_labels",
		mcp.WithDescription("List all labels (tags) used to organize devices and synthetic tests. Labels provide logical grouping — for example, 'inet' for internet-facing routers, 'emea' for European devices. Device labels appear in device details and can be used for filtering."),
	), makeListLabelsHandler(c))

	s.AddTool(mcp.NewTool("get_label",
		mcp.WithDescription("Get details of a specific label including name, description, color, and which devices/tests use it."),
		mcp.WithString("label_id", mcp.Required(), mcp.Description("The label ID")),
	), makeGetLabelHandler(c))
}

func makeListLabelsHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := c.V6("GET", "/label/v202210/labels", nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list labels: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}

func makeGetLabelHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		labelID, err := request.RequireString("label_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := c.V6("GET", fmt.Sprintf("/label/v202210/labels/%s", labelID), nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get label: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}
