package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerAlertingTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("list_alert_policies",
		mcp.WithDescription("List DDoS protection and alerting policies. Policies define detection rules with: dimensions (what to group by, e.g. IP_dst), metrics (what to measure, e.g. packets, bits, unique_src_ip), filters (what traffic to watch), and thresholds (severity levels with conditions). Filter by status: 'active' (monitoring), 'disabled' (paused), 'error' (broken config). Use get_alert_policy for threshold details."),
		mcp.WithString("status",
			mcp.Description("Filter by status: 'active', 'disabled', 'error', or leave empty for all")),
	), makeListAlertPoliciesHandler(c))

	s.AddTool(mcp.NewTool("get_alert_policy",
		mcp.WithDescription("Get full configuration of an alert policy including threshold conditions (e.g., packets >= 500K triggers minor, >= 8M triggers critical), traffic filters (which prefixes/protocols are monitored), and notification settings. Essential for understanding DDoS detection sensitivity and coverage gaps."),
		mcp.WithString("policy_id", mcp.Required(),
			mcp.Description("The alert policy ID (from list_alert_policies)")),
	), makeGetAlertPolicyHandler(c))

	s.AddTool(mcp.NewTool("list_active_alarms",
		mcp.WithDescription("List currently active (firing) alarms. Returns ongoing threshold violations including DDoS attacks in progress, traffic anomalies, and capacity alerts. If empty, the network is clean. Use list_alert_history for past events."),
		mcp.WithNumber("lookback_minutes",
			mcp.Description("How far back to look for active alarms (default: 60 = last hour, use 1440 for last 24h)")),
	), makeListActiveAlarmsHandler(c))

	s.AddTool(mcp.NewTool("list_alert_history",
		mcp.WithDescription("List historical alert events — past DDoS attacks, threshold violations, and their resolution. Shows alarm start/end times, policy name, target (IP/prefix), severity (minor/major/critical), state (CLEAR/ALARM), and peak value. Use top_ddos_attacks for a ranked summary. API retains ~30 days of history."),
		mcp.WithNumber("lookback_minutes",
			mcp.Description("How far back to look (default: 10080 = 7 days, max: 43200 = 30 days)")),
	), makeListAlertHistoryHandler(c))

	s.AddTool(mcp.NewTool("top_ddos_attacks",
		mcp.WithDescription("Rank the largest DDoS attacks by volume (bps) or packet rate (pps). Returns a sorted table with date, target, peak value, severity, duration, and triggering policy. Use sort_by='bps' for bandwidth-based attacks, sort_by='pps' for packet floods. For forensic analysis of a specific attack, use query_flows with starting_time/ending_time and a filter on the target IP."),
		mcp.WithNumber("top",
			mcp.Description("Number of top attacks to return (default: 5)")),
		mcp.WithNumber("lookback_minutes",
			mcp.Description("How far back to look (default: 43200 = 30 days)")),
		mcp.WithString("sort_by",
			mcp.Description("Sort by 'bps' for volume or 'pps' for packet rate (default: bps)")),
	), makeTopDDoSHandler(c))
}

func makeListAlertPoliciesHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := c.V5("GET", "/alerting/policies", nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list policies: %v", err)), nil
		}

		resp, err := parsePolicies(data)
		if err != nil {
			return mcp.NewToolResultText(formatJSON(data)), nil
		}

		statusFilter := request.GetString("status", "")
		policies := resp.Policies
		if statusFilter != "" {
			statusFilter = strings.ToLower(statusFilter)
			var filtered []map[string]any
			for _, p := range policies {
				if strings.ToLower(fmt.Sprintf("%v", p["status"])) == statusFilter {
					filtered = append(filtered, p)
				}
			}
			policies = filtered
		}

		if len(policies) == 0 {
			return mcp.NewToolResultText("No alert policies found matching the filter."), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("## Alert Policies (%d of %d total)\n\n", len(policies), resp.Count))
		sb.WriteString("| # | ID | Name | Status | Dimensions | Metrics |\n")
		sb.WriteString("|---|-----|------|--------|-----------|--------|\n")

		for i, p := range policies {
			dims := joinAny(p["dimensions"])
			metrics := joinAny(p["metrics"])
			name := fmt.Sprintf("%v", p["name"])
			if len(name) > 55 {
				name = name[:55] + "..."
			}
			sb.WriteString(fmt.Sprintf("| %d | %v | %s | %v | %s | %s |\n",
				i+1, p["id"], name, p["status"], dims, metrics))
		}

		return mcp.NewToolResultText(sb.String()), nil
	}
}

func makeGetAlertPolicyHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		policyID, err := request.RequireString("policy_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, err := c.V5("GET", "/alerting/policies", nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch policies: %v", err)), nil
		}

		resp, err := parsePolicies(data)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to parse policies: %v", err)), nil
		}

		for _, p := range resp.Policies {
			if fmt.Sprintf("%v", p["id"]) == policyID {
				pretty, _ := json.MarshalIndent(p, "", "  ")
				return mcp.NewToolResultText(summarizePolicy(p) + "\n<details><summary>Full JSON</summary>\n\n```json\n" + string(pretty) + "\n```\n</details>\n"), nil
			}
		}

		return mcp.NewToolResultError(fmt.Sprintf("Policy %s not found", policyID)), nil
	}
}

func makeListActiveAlarmsHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		lookback := request.GetInt("lookback_minutes", 60)
		path := fmt.Sprintf("/alerts-active/alarms?lookback_minutes=%d", lookback)

		data, err := c.V5("GET", path, nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list alarms: %v", err)), nil
		}

		var alarms []map[string]any
		if err := json.Unmarshal(data, &alarms); err != nil {
			return mcp.NewToolResultText(formatJSON(data)), nil
		}

		if len(alarms) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No active alarms in the last %d minutes. Network is clean.", lookback)), nil
		}

		return mcp.NewToolResultText(summarizeAlarms(alarms, "Active Alarms")), nil
	}
}

func makeListAlertHistoryHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		lookback := request.GetInt("lookback_minutes", 10080)
		path := fmt.Sprintf("/alerts-active/alerts-history?lookback_minutes=%d", lookback)

		data, err := c.V5("GET", path, nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list alert history: %v", err)), nil
		}

		var events []map[string]any
		if err := json.Unmarshal(data, &events); err != nil {
			return mcp.NewToolResultText(formatJSON(data)), nil
		}

		if len(events) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No alert events in the last %d minutes.", lookback)), nil
		}

		return mcp.NewToolResultText(summarizeAlarms(events, "Alert History")), nil
	}
}

func makeTopDDoSHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		top := request.GetInt("top", 5)
		lookback := request.GetInt("lookback_minutes", 43200)
		sortBy := request.GetString("sort_by", "bps")

		path := fmt.Sprintf("/alerts-active/alerts-history?lookback_minutes=%d", lookback)
		data, err := c.V5("GET", path, nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch alert history: %v", err)), nil
		}

		var events []map[string]any
		if err := json.Unmarshal(data, &events); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to parse events: %v", err)), nil
		}

		if len(events) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No DDoS events in the last %d minutes.", lookback)), nil
		}

		// Build scored list
		type scored struct {
			event map[string]any
			value float64
		}
		var candidates []scored

		for _, e := range events {
			metrics := joinAny(e["alert_metric"])
			val, _ := e["alert_value"].(float64)
			val2, _ := e["alert_value2nd"].(float64)
			val3, _ := e["alert_value3rd"].(float64)

			// Map metric positions to values
			metricList := strings.Split(metrics, ", ")
			vals := []float64{val, val2, val3}

			// Pick the right value based on sort preference
			target := "bits"
			if sortBy == "pps" {
				target = "packets"
			}

			var pick float64
			found := false
			for idx, m := range metricList {
				if m == target && idx < len(vals) {
					pick = vals[idx]
					found = true
					break
				}
			}
			if !found {
				// This event doesn't track the requested metric — skip it
				continue
			}

			if pick > 0 {
				candidates = append(candidates, scored{event: e, value: pick})
			}
		}

		// Sort descending
		for i := 0; i < len(candidates); i++ {
			for j := i + 1; j < len(candidates); j++ {
				if candidates[j].value > candidates[i].value {
					candidates[i], candidates[j] = candidates[j], candidates[i]
				}
			}
		}

		if top > len(candidates) {
			top = len(candidates)
		}

		unit := "bps"
		if sortBy == "pps" {
			unit = "pps"
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("## Top %d DDoS Attacks by %s (Last %d days)\n\n", top, strings.ToUpper(unit), lookback/1440))

		sb.WriteString("| # | Date | Target | Peak | Severity | Duration | Policy |\n")
		sb.WriteString("|---|------|--------|------|----------|----------|--------|\n")

		for i, c := range candidates[:top] {
			e := c.event
			start := fmt.Sprintf("%v", e["alarm_start"])
			if len(start) > 16 {
				start = start[:16]
			}
			target := fmt.Sprintf("%v", e["alert_key"])
			if lookup, ok := e["alert_key_lookup"].(string); ok && lookup != "" {
				target = lookup
			}
			severity := fmt.Sprintf("%v", e["alert_severity"])
			policy := fmt.Sprintf("%v", e["policy_name"])
			if len(policy) > 45 {
				policy = policy[:45] + "..."
			}

			// Format peak value
			var peak string
			if sortBy == "pps" {
				peak = formatPPS(c.value)
			} else {
				peak = formatRate(c.value, "bytes")
			}

			// Duration
			duration := "?"
			startTime := fmt.Sprintf("%v", e["alarm_start"])
			endTime := fmt.Sprintf("%v", e["alarm_end"])
			if len(startTime) >= 19 && len(endTime) >= 19 {
				// Simple duration calc from ISO timestamps
				duration = calcDuration(startTime, endTime)
			}

			sb.WriteString(fmt.Sprintf("| %d | %s | %s | **%s** | %s | %s | %s |\n",
				i+1, start, target, peak, severity, duration, policy))
		}

		return mcp.NewToolResultText(sb.String()), nil
	}
}

func formatPPS(v float64) string {
	switch {
	case v >= 1e6:
		return fmt.Sprintf("%.2fM pps", v/1e6)
	case v >= 1e3:
		return fmt.Sprintf("%.1fK pps", v/1e3)
	default:
		return fmt.Sprintf("%.0f pps", v)
	}
}

func calcDuration(start, end string) string {
	// Parse "2026-03-25T13:14:07..." format
	if len(start) < 19 || len(end) < 19 {
		return "?"
	}
	parseTime := func(s string) (int, int) {
		// Extract hours and minutes
		h := 0
		m := 0
		fmt.Sscanf(s[11:16], "%d:%d", &h, &m)
		// Extract day
		d := 0
		fmt.Sscanf(s[8:10], "%d", &d)
		return d*24*60 + h*60 + m, 0
	}
	startMin, _ := parseTime(start)
	endMin, _ := parseTime(end)
	diff := endMin - startMin
	if diff < 0 {
		diff += 24 * 60 // wrap around midnight
	}
	if diff < 60 {
		return fmt.Sprintf("%dm", diff)
	}
	return fmt.Sprintf("%dh%dm", diff/60, diff%60)
}

func summarizePolicy(p map[string]any) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Policy: %v\n\n", p["name"]))
	sb.WriteString(fmt.Sprintf("- **ID:** %v\n", p["id"]))
	sb.WriteString(fmt.Sprintf("- **Status:** %v\n", p["status"]))
	sb.WriteString(fmt.Sprintf("- **Dimensions:** %s\n", joinAny(p["dimensions"])))
	sb.WriteString(fmt.Sprintf("- **Metrics:** %s\n", joinAny(p["metrics"])))

	if filters, ok := p["filters"].(map[string]any); ok {
		sb.WriteString(fmt.Sprintf("- **Filter connector:** %v\n", filters["connector"]))
		if groups, ok := filters["filterGroups"].([]any); ok {
			for gi, g := range groups {
				if group, ok := g.(map[string]any); ok {
					if fs, ok := group["filters"].([]any); ok {
						for _, f := range fs {
							if filter, ok := f.(map[string]any); ok {
								sb.WriteString(fmt.Sprintf("  - Group %d: %v %v %v\n",
									gi+1, filter["filterField"], filter["operator"], filter["filterValue"]))
							}
						}
					}
				}
			}
		}
	}

	if thresholds, ok := p["thresholds"].([]any); ok {
		sb.WriteString("\n### Thresholds\n\n")
		for _, t := range thresholds {
			if th, ok := t.(map[string]any); ok {
				sb.WriteString(fmt.Sprintf("**Severity: %v**\n", th["severity"]))
				if conditions, ok := th["conditions"].([]any); ok {
					for _, c := range conditions {
						if cond, ok := c.(map[string]any); ok {
							sb.WriteString(fmt.Sprintf("- %v %v %v (type: %v)\n",
								cond["metric"], cond["operator"], cond["comparisonValue"], cond["type"]))
						}
					}
				}
			}
		}
	}

	return sb.String()
}

func summarizeAlarms(alarms []map[string]any, title string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## %s (%d events)\n\n", title, len(alarms)))
	sb.WriteString("| Time | Policy | Target | Severity | State | Value |\n")
	sb.WriteString("|------|--------|--------|----------|-------|-------|\n")

	for _, a := range alarms {
		start := fmt.Sprintf("%v", a["alarm_start"])
		if len(start) > 16 {
			start = start[:16]
		}
		policy := fmt.Sprintf("%v", a["policy_name"])
		if policy == "<nil>" {
			policy = fmt.Sprintf("ID:%v", a["alert_id"])
		}
		if len(policy) > 40 {
			policy = policy[:40] + "..."
		}
		target := fmt.Sprintf("%v", a["alert_key"])
		if lookup, ok := a["alert_key_lookup"].(string); ok && lookup != "" {
			target = lookup
		}
		severity := fmt.Sprintf("%v", a["alert_severity"])
		state := fmt.Sprintf("%v", a["alarm_state"])
		if newState, ok := a["new_alarm_state"].(string); ok && newState != "" {
			state = newState
		}
		value := formatAlertValue(a)
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			start, policy, target, severity, state, value))
	}

	return sb.String()
}

func formatAlertValue(a map[string]any) string {
	v, ok := a["alert_value"].(float64)
	if !ok {
		return "N/A"
	}
	metrics := joinAny(a["alert_metric"])
	if strings.Contains(metrics, "bits") {
		return formatRate(v, "bytes")
	}
	if strings.Contains(metrics, "packets") {
		return formatRate(v, "packets")
	}
	return fmt.Sprintf("%.0f", v)
}

type policiesResponse struct {
	Policies []map[string]any
	Count    int
}

func parsePolicies(data json.RawMessage) (*policiesResponse, error) {
	// Use json.Decoder with UseNumber to handle large integers
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.UseNumber()
	var raw map[string]any
	if err := dec.Decode(&raw); err != nil {
		return nil, err
	}

	resp := &policiesResponse{}
	if count, ok := raw["count"].(json.Number); ok {
		c, _ := count.Int64()
		resp.Count = int(c)
	}

	if policies, ok := raw["policies"].([]any); ok {
		for _, p := range policies {
			if pm, ok := p.(map[string]any); ok {
				resp.Policies = append(resp.Policies, pm)
			}
		}
	}

	return resp, nil
}

func joinAny(v any) string {
	if arr, ok := v.([]any); ok {
		parts := make([]string, len(arr))
		for i, item := range arr {
			parts[i] = fmt.Sprintf("%v", item)
		}
		return strings.Join(parts, ", ")
	}
	if arr, ok := v.([]string); ok {
		return strings.Join(arr, ", ")
	}
	return fmt.Sprintf("%v", v)
}
