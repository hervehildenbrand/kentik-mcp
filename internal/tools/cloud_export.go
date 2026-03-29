package tools

import (
	"context"
	"fmt"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerCloudExportTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("list_cloud_exports",
		mcp.WithDescription("List cloud flow log exports configured in Kentik. Cloud exports ingest VPC flow logs from AWS, Azure, GCP, or Oracle Cloud. Each export shows the cloud provider, account/subscription, region, and status. Flow data from cloud exports is queryable via query_flows like router flow data."),
	), makeListCloudExportsHandler(c))

	s.AddTool(mcp.NewTool("get_cloud_export",
		mcp.WithDescription("Get detailed configuration of a cloud export including provider-specific settings (AWS IAM role, Azure subscription, GCP project, etc.)."),
		mcp.WithString("export_id", mcp.Required(),
			mcp.Description("The cloud export ID")),
	), makeGetCloudExportHandler(c))
}

func makeListCloudExportsHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := c.V6("GET", "/cloud_export/v202101beta1/exports", nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list cloud exports: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}

func makeGetCloudExportHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		exportID, err := request.RequireString("export_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := c.V6("GET", fmt.Sprintf("/cloud_export/v202101beta1/exports/%s", exportID), nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get cloud export: %v", err)), nil
		}
		return mcp.NewToolResultText(formatJSON(data)), nil
	}
}
