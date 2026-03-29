package tools

import (
	"context"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerQueryTool(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("kentik_query",
		mcp.WithDescription("Query Kentik network flow data. Actions: "+
			"'flows' — top-X analysis by dimension (AS, IP, port, protocol, tcp_flags, geography, interface). "+
			"'time_series' — avg/p95/max trends over time for capacity planning. "+
			"'top_ddos' — rank DDoS attacks by volume (bps) or packet rate (pps). "+
			"Common dimensions: AS_src, AS_dst, IP_src, IP_dst, Port_src, Port_dst, Proto, tcp_flags, "+
			"Geography_src, Geography_dst, InterfaceID_src, InterfaceID_dst, i_device_site_name, TopFlow. "+
			"Common filter fields: dst_as, src_as, inet_dst_addr, inet_src_addr, l4_dst_port, l4_src_port, protocol, tcp_flags. "+
			"Filter format: {\"filterGroups\":[{\"filters\":[{\"filterField\":\"dst_as\",\"operator\":\"=\",\"filterValue\":\"15169\"}]}]}"),
		mcp.WithString("action", mcp.Required(),
			mcp.Description("'flows', 'time_series', or 'top_ddos'")),
		mcp.WithString("metric",
			mcp.Description("Traffic metric: bytes, in_bytes, out_bytes, packets, in_packets, out_packets, fps, unique_src_ip, unique_dst_ip (required for flows/time_series)")),
		mcp.WithString("dimension",
			mcp.Description("Group-by dimension(s), comma-separated (required for flows/time_series)")),
		mcp.WithNumber("lookback_seconds",
			mcp.Description("Look-back in seconds. Default 3600=1h. Set to 0 for absolute time range.")),
		mcp.WithString("starting_time",
			mcp.Description("Absolute start 'YYYY-MM-DD HH:mm' UTC. Requires lookback_seconds=0.")),
		mcp.WithString("ending_time",
			mcp.Description("Absolute end 'YYYY-MM-DD HH:mm' UTC.")),
		mcp.WithNumber("topx",
			mcp.Description("Number of top results, 1-40 (default 10)")),
		mcp.WithString("device_name",
			mcp.Description("Comma-delimited device names to filter")),
		mcp.WithString("filters_json",
			mcp.Description("Raw JSON filter object")),
		mcp.WithString("sort_by",
			mcp.Description("For top_ddos: 'bps' or 'pps' (default 'bps')")),
		mcp.WithNumber("top",
			mcp.Description("For top_ddos: number of results (default 5)")),
		mcp.WithNumber("lookback_minutes",
			mcp.Description("For top_ddos: lookback in minutes (default 43200=30d)")),
	), makeQueryRouter(c))
}

func makeQueryRouter(c *client.Client) server.ToolHandlerFunc {
	flowsHandler := makeQueryFlowsHandler(c)
	timeSeriesHandler := makeQueryTimeSeriesHandler(c)
	topDDoSHandler := makeTopDDoSHandler(c)

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		action, err := request.RequireString("action")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		switch action {
		case "flows":
			return flowsHandler(ctx, request)
		case "time_series":
			return timeSeriesHandler(ctx, request)
		case "top_ddos":
			return topDDoSHandler(ctx, request)
		default:
			return mcp.NewToolResultError("Unknown action: " + action + ". Use 'flows', 'time_series', or 'top_ddos'."), nil
		}
	}
}
