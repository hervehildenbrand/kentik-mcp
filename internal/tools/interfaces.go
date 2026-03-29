package tools

import (
	"context"
	"fmt"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerInterfaceTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("list_interfaces",
		mcp.WithDescription("List all interfaces on a specific device. Returns SNMP ID, description, speed, connectivity type (transit, ix, backbone, pni, customer), network boundary (internal/external), and operational status. Use interface data to understand link utilization — combine with query_flows dimension 'InterfaceID_src' or 'InterfaceID_dst' for per-interface traffic."),
		mcp.WithString("device_id", mcp.Required(), mcp.Description("The device ID (from list_devices)")),
	), makeListInterfacesHandler(c))

	s.AddTool(mcp.NewTool("get_interface",
		mcp.WithDescription("Get detailed information about a specific interface including SNMP metrics (MTU, speed, oper status), connectivity type, provider assignment, VRF, and interface tags. Note: interface_id is the Kentik internal ID (e.g. '8188322'), not the SNMP index."),
		mcp.WithString("device_id", mcp.Required(), mcp.Description("The device ID")),
		mcp.WithString("interface_id", mcp.Required(), mcp.Description("The Kentik interface ID (from list_interfaces 'id' field, not snmp_id)")),
	), makeGetInterfaceHandler(c))
}

func makeListInterfacesHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		deviceID, err := request.RequireString("device_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := c.V5("GET", fmt.Sprintf("/device/%s/interfaces", deviceID), nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list interfaces: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}

func makeGetInterfaceHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		deviceID, err := request.RequireString("device_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		interfaceID, err := request.RequireString("interface_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := c.V5("GET", fmt.Sprintf("/device/%s/interface/%s", deviceID, interfaceID), nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get interface: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}
