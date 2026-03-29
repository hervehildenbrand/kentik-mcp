package tools

import (
	"context"
	"fmt"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerDeviceTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("list_devices",
		mcp.WithDescription("List all network devices (routers, switches) registered in Kentik. Each device has an ID, name, type, site, BGP configuration, sample rate, and labels. Use device names in query_flows 'device_name' parameter to filter traffic. Use device IDs with get_device, list_interfaces. Related: list_sites shows device locations, list_labels shows device groupings."),
	), makeListDevicesHandler(c))

	s.AddTool(mcp.NewTool("get_device",
		mcp.WithDescription("Get full details of a specific device including BGP neighbor config (IP, ASN, peer addresses), sending IPs, SNMP settings, flow type, sample rate, site assignment, and applied labels."),
		mcp.WithString("device_id", mcp.Required(), mcp.Description("The device ID (numeric string, e.g. '4931'). Find IDs via list_devices.")),
	), makeGetDeviceHandler(c))
}

func makeListDevicesHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := c.V5("GET", "/devices", nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list devices: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}

func makeGetDeviceHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		deviceID, err := request.RequireString("device_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := c.V5("GET", fmt.Sprintf("/device/%s", deviceID), nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get device: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}
