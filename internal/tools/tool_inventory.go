package tools

import (
	"context"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerInventoryTool(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("kentik_inventory",
		mcp.WithDescription("Query Kentik network infrastructure inventory. Actions: "+
			"'list_devices' — all routers/switches with site, BGP config, labels. "+
			"'get_device' — device detail (requires device_id). "+
			"'list_interfaces' — interfaces on a device with speed, connectivity type (requires device_id). "+
			"'get_interface' — interface detail (requires device_id + interface_id). "+
			"'list_sites' — physical locations (data centers, PoPs). "+
			"'get_site' — site detail with coordinates (requires site_id). "+
			"'list_labels' — device/test labels for grouping. "+
			"'get_label' — label detail (requires label_id)."),
		mcp.WithString("action", mcp.Required(),
			mcp.Description("One of: list_devices, get_device, list_interfaces, get_interface, list_sites, get_site, list_labels, get_label")),
		mcp.WithString("device_id", mcp.Description("Device ID for get_device, list_interfaces, get_interface")),
		mcp.WithString("interface_id", mcp.Description("Interface ID for get_interface (Kentik internal ID, not SNMP index)")),
		mcp.WithString("site_id", mcp.Description("Site ID for get_site")),
		mcp.WithString("label_id", mcp.Description("Label ID for get_label")),
	), makeInventoryRouter(c))
}

func makeInventoryRouter(c *client.Client) server.ToolHandlerFunc {
	listDevices := makeListDevicesHandler(c)
	getDevice := makeGetDeviceHandler(c)
	listInterfaces := makeListInterfacesHandler(c)
	getInterface := makeGetInterfaceHandler(c)
	listSites := makeListSitesHandler(c)
	getSite := makeGetSiteHandler(c)
	listLabels := makeListLabelsHandler(c)
	getLabel := makeGetLabelHandler(c)

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		action, err := request.RequireString("action")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		switch action {
		case "list_devices":
			return listDevices(ctx, request)
		case "get_device":
			return getDevice(ctx, request)
		case "list_interfaces":
			return listInterfaces(ctx, request)
		case "get_interface":
			return getInterface(ctx, request)
		case "list_sites":
			return listSites(ctx, request)
		case "get_site":
			return getSite(ctx, request)
		case "list_labels":
			return listLabels(ctx, request)
		case "get_label":
			return getLabel(ctx, request)
		default:
			return mcp.NewToolResultError("Unknown action: " + action), nil
		}
	}
}
