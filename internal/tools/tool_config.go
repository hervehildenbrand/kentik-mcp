package tools

import (
	"context"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerConfigTool(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("kentik_config",
		mcp.WithDescription("Query Kentik configuration and admin objects. Actions: "+
			"'list_saved_filters' / 'get_saved_filter' — reusable query filter definitions (requires filter_id). "+
			"'list_custom_dimensions' / 'get_custom_dimension' — flow enrichment dimensions with populator rules (requires dimension_id). "+
			"'list_users' / 'get_user' — organization users and roles (requires user_id). "+
			"'list_cloud_exports' / 'get_cloud_export' — AWS/Azure/GCP/OCI flow log imports (requires export_id). "+
			"'list_notification_channels' / 'get_notification_channel' — alert destinations (email, Slack, PagerDuty) (requires channel_id)."),
		mcp.WithString("action", mcp.Required(),
			mcp.Description("One of: list_saved_filters, get_saved_filter, list_custom_dimensions, get_custom_dimension, list_users, get_user, list_cloud_exports, get_cloud_export, list_notification_channels, get_notification_channel")),
		mcp.WithString("id", mcp.Description("Object ID for any get_ action (filter_id, dimension_id, user_id, export_id, or channel_id)")),
	), makeConfigRouter(c))
}

func makeConfigRouter(c *client.Client) server.ToolHandlerFunc {
	listFilters := makeListSavedFiltersHandler(c)
	getFilter := makeGetSavedFilterHandler(c)
	listDims := makeListCustomDimensionsHandler(c)
	getDim := makeGetCustomDimensionHandler(c)
	listUsers := makeListUsersHandler(c)
	getUser := makeGetUserHandler(c)
	listExports := makeListCloudExportsHandler(c)
	getExport := makeGetCloudExportHandler(c)
	listChannels := makeListNotificationChannelsHandler(c)
	getChannel := makeGetNotificationChannelHandler(c)

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		action, err := request.RequireString("action")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		switch action {
		case "list_saved_filters":
			return listFilters(ctx, request)
		case "get_saved_filter":
			return remapID(getFilter, "filter_id")(ctx, request)
		case "list_custom_dimensions":
			return listDims(ctx, request)
		case "get_custom_dimension":
			return remapID(getDim, "dimension_id")(ctx, request)
		case "list_users":
			return listUsers(ctx, request)
		case "get_user":
			return remapID(getUser, "user_id")(ctx, request)
		case "list_cloud_exports":
			return listExports(ctx, request)
		case "get_cloud_export":
			return remapID(getExport, "export_id")(ctx, request)
		case "list_notification_channels":
			return listChannels(ctx, request)
		case "get_notification_channel":
			return remapID(getChannel, "channel_id")(ctx, request)
		default:
			return mcp.NewToolResultError("Unknown action: " + action), nil
		}
	}
}

// remapID creates a wrapper handler that copies the "id" parameter to the target key
// so that underlying handlers expecting specific parameter names (filter_id, user_id, etc.) work correctly.
func remapID(handler server.ToolHandlerFunc, targetKey string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if id := request.GetString("id", ""); id != "" {
			args := request.GetArguments()
			if args == nil {
				args = map[string]any{}
			}
			args[targetKey] = id
			request.Params.Arguments = args
		}
		return handler(ctx, request)
	}
}
