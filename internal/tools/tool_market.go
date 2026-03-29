package tools

import (
	"context"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerMarketTool(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("kentik_market",
		mcp.WithDescription("Kentik Market Intelligence (KMI) — look up AS information and relationships. Actions: "+
			"'as_info' — AS details: name, country, provider/customer/peer counts. "+
			"'as_relationships' — transit providers, customers, or peers of an AS (requires relationship_type). "+
			"Note: KMI may not be available on all Kentik regions."),
		mcp.WithString("action", mcp.Required(),
			mcp.Description("One of: as_info, as_relationships")),
		mcp.WithString("asn", mcp.Required(),
			mcp.Description("AS number (e.g. '15169' for Google)")),
		mcp.WithString("relationship_type",
			mcp.Description("For as_relationships: 'customers', 'providers', or 'peers'")),
	), makeMarketRouter(c))
}

func makeMarketRouter(c *client.Client) server.ToolHandlerFunc {
	asInfo := makeGetASInfoHandler(c)
	asRel := makeGetASRelationshipsHandler(c)

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		action, err := request.RequireString("action")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		switch action {
		case "as_info":
			return asInfo(ctx, request)
		case "as_relationships":
			return asRel(ctx, request)
		default:
			return mcp.NewToolResultError("Unknown action: " + action), nil
		}
	}
}
