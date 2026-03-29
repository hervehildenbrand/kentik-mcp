package tools

import (
	"context"
	"fmt"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerNotificationTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("list_notification_channels",
		mcp.WithDescription("List notification channels where alert and synthetic test notifications are delivered. Channels can be email, Slack, PagerDuty, OpsGenie, ServiceNow, or generic webhooks. Alert policies and synthetic tests reference these channels to send notifications when thresholds are violated."),
	), makeListNotificationChannelsHandler(c))

	s.AddTool(mcp.NewTool("get_notification_channel",
		mcp.WithDescription("Get detailed configuration of a notification channel including type, destination, and linked policies."),
		mcp.WithString("channel_id", mcp.Required(),
			mcp.Description("The notification channel ID")),
	), makeGetNotificationChannelHandler(c))
}

func makeListNotificationChannelsHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := c.V6("GET", "/notification_channel/v202204beta1/channels", nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list notification channels: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}

func makeGetNotificationChannelHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		channelID, err := request.RequireString("channel_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := c.V6("GET", fmt.Sprintf("/notification_channel/v202204beta1/channels/%s", channelID), nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get notification channel: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}
