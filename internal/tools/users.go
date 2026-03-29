package tools

import (
	"context"
	"fmt"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerUserTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("list_users",
		mcp.WithDescription("List all users in your Kentik organization. Shows user email, full name, role (Member, Administrator, Super Administrator), and status. Useful for auditing API access and understanding team structure."),
	), makeListUsersHandler(c))

	s.AddTool(mcp.NewTool("get_user",
		mcp.WithDescription("Get detailed information about a specific user including email, role, permissions, and last login time."),
		mcp.WithString("user_id", mcp.Required(), mcp.Description("The user ID")),
	), makeGetUserHandler(c))
}

func makeListUsersHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := c.V5("GET", "/users", nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list users: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}

func makeGetUserHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID, err := request.RequireString("user_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := c.V5("GET", fmt.Sprintf("/user/%s", userID), nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get user: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}
