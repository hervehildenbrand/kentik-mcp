package tools

import (
	"context"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerAlertsTool(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("kentik_alerts",
		mcp.WithDescription("Query Kentik DDoS protection and alerting system. Actions: "+
			"'list_policies' — DDoS detection policies (optional status filter: active/disabled/error). "+
			"'get_policy' — policy detail with thresholds and conditions (requires policy_id). "+
			"'active_alarms' — currently firing alarms and attacks. "+
			"'alert_history' — past alert events, ~30 day retention. "+
			"'top_attacks' — rank largest DDoS attacks by bps or pps."),
		mcp.WithString("action", mcp.Required(),
			mcp.Description("One of: list_policies, get_policy, active_alarms, alert_history, top_attacks")),
		mcp.WithString("policy_id", mcp.Description("Policy ID for get_policy")),
		mcp.WithString("status", mcp.Description("Filter for list_policies: 'active', 'disabled', 'error'")),
		mcp.WithNumber("lookback_minutes", mcp.Description("Lookback: active_alarms default 60, alert_history/top_attacks default 43200=30d")),
		mcp.WithString("sort_by", mcp.Description("For top_attacks: 'bps' or 'pps' (default 'bps')")),
		mcp.WithNumber("top", mcp.Description("For top_attacks: number of results (default 5)")),
	), makeAlertsRouter(c))
}

func makeAlertsRouter(c *client.Client) server.ToolHandlerFunc {
	listPolicies := makeListAlertPoliciesHandler(c)
	getPolicy := makeGetAlertPolicyHandler(c)
	activeAlarms := makeListActiveAlarmsHandler(c)
	alertHistory := makeListAlertHistoryHandler(c)
	topAttacks := makeTopDDoSHandler(c)

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		action, err := request.RequireString("action")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		switch action {
		case "list_policies":
			return listPolicies(ctx, request)
		case "get_policy":
			return getPolicy(ctx, request)
		case "active_alarms":
			return activeAlarms(ctx, request)
		case "alert_history":
			return alertHistory(ctx, request)
		case "top_attacks":
			return topAttacks(ctx, request)
		default:
			return mcp.NewToolResultError("Unknown action: " + action), nil
		}
	}
}
