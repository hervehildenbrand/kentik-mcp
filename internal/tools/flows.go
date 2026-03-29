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

func registerFlowTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("query_flows",
		mcp.WithDescription("Query Kentik network flow data (top-X). The primary tool for traffic analysis — returns ranked results grouped by dimension(s). "+
			"Supports multiple dimensions (comma-separated) for cross-tabulation. Results include avg, p95, and max rates. "+
			"Common dimensions: AS_src, AS_dst, IP_src, IP_dst, Port_src, Port_dst, Proto, tcp_flags, "+
			"Geography_src, Geography_dst, InterfaceID_src, InterfaceID_dst, i_device_site_name, i_device_name, "+
			"IP_dst_cidr_24_64, TopFlow. "+
			"Common filter fields for filters_json: dst_as, src_as, inet_dst_addr, inet_src_addr, l4_dst_port, "+
			"l4_src_port, protocol, tcp_flags, i_device_name, i_src_network_bndry_name, i_output_snmp_alias, i_trf_profile. "+
			"Filter operators: =, !=, contains, not_contains. "+
			"Use starting_time/ending_time for historical forensics (e.g., investigating a past DDoS attack)."),
		mcp.WithString("metric", mcp.Required(),
			mcp.Description("Traffic metric: bytes, in_bytes, out_bytes, packets, in_packets, out_packets, fps (flows/sec), unique_src_ip, unique_dst_ip")),
		mcp.WithString("dimension", mcp.Required(),
			mcp.Description("Group-by dimension(s), comma-separated. Can combine multiple: 'AS_src,Port_dst' or 'tcp_flags,Port_dst'")),
		mcp.WithNumber("lookback_seconds",
			mcp.Description("Look-back time in seconds (default 3600 = last hour). Set to 0 when using starting_time/ending_time.")),
		mcp.WithString("starting_time",
			mcp.Description("Absolute start time in 'YYYY-MM-DD HH:mm' format UTC (e.g. '2026-03-02 21:20'). Requires lookback_seconds=0.")),
		mcp.WithString("ending_time",
			mcp.Description("Absolute end time in 'YYYY-MM-DD HH:mm' format UTC. Requires lookback_seconds=0.")),
		mcp.WithNumber("topx",
			mcp.Description("Number of top results to return, 1-40 (default 10)")),
		mcp.WithString("device_name",
			mcp.Description("Comma-delimited list of device names to filter by (from list_devices). Omit to query all devices.")),
		mcp.WithString("filters_json",
			mcp.Description("Raw JSON filter object. Format: {\"filterGroups\":[{\"filters\":[{\"filterField\":\"dst_as\",\"operator\":\"=\",\"filterValue\":\"15169\"}]}]}. Common fields: dst_as, src_as, inet_dst_addr, inet_src_addr, l4_dst_port, l4_src_port, protocol, tcp_flags.")),
	), makeQueryFlowsHandler(c))

	s.AddTool(mcp.NewTool("query_time_series",
		mcp.WithDescription("Query Kentik flow data with time series statistics (avg, p95, max) over a time window. "+
			"Returns the same dimensions as query_flows but formatted as a trend table. "+
			"Best for capacity planning (7-day P95 by site), trend comparison, and understanding traffic patterns over time. "+
			"Use for questions like 'what was the peak traffic this week?' or '24h trend by destination AS'."),
		mcp.WithString("metric", mcp.Required(),
			mcp.Description("Traffic metric: bytes, packets, fps")),
		mcp.WithString("dimension", mcp.Required(),
			mcp.Description("Group-by dimension(s), comma-separated. Same dimensions as query_flows.")),
		mcp.WithNumber("lookback_seconds",
			mcp.Description("Look-back time in seconds (default 3600). Common: 86400=24h, 604800=7d")),
		mcp.WithNumber("topx",
			mcp.Description("Number of top results (default 10)")),
	), makeQueryTimeSeriesHandler(c))
}

func makeQueryFlowsHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		metric, err := request.RequireString("metric")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		dimensionStr, err := request.RequireString("dimension")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		dimensions := parseDimensions(dimensionStr)
		lookback := request.GetInt("lookback_seconds", 3600)
		topx := request.GetInt("topx", 10)
		outsort := getOutsort(metric)

		deviceName := request.GetString("device_name", "")
		if deviceName == "" {
			names, err := c.ListDeviceNames()
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch devices: %v", err)), nil
			}
			deviceName = strings.Join(names, ",")
		}

		query := map[string]any{
			"metric":           metric,
			"dimension":        dimensions,
			"topx":             topx,
			"depth":            100,
			"fastData":         "Auto",
			"outsort":          outsort,
			"lookback_seconds": lookback,
			"time_format":      "UTC",
			"hostname_lookup":  true,
			"device_name":      deviceName,
			"all_selected":     false,
		}

		// Absolute time range overrides lookback
		if startTime := request.GetString("starting_time", ""); startTime != "" {
			query["starting_time"] = startTime
			query["ending_time"] = request.GetString("ending_time", "")
			query["lookback_seconds"] = 0
		}

		if filtersJSON := request.GetString("filters_json", ""); filtersJSON != "" {
			var filters map[string]any
			if err := json.Unmarshal([]byte(filtersJSON), &filters); err == nil {
				query["filters_obj"] = filters
			}
		}

		body := map[string]any{
			"queries": []map[string]any{
				{
					"query":       query,
					"bucket":      "Left +Y Axis",
					"bucketIndex": 0,
					"isOverlay":   false,
				},
			},
		}

		data, err := c.V5("POST", "/query/topXdata", body)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to query flows: %v", err)), nil
		}

		return mcp.NewToolResultText(summarizeFlowResults(data, metric)), nil
	}
}

func makeQueryTimeSeriesHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		metric, err := request.RequireString("metric")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		dimensionStr, err := request.RequireString("dimension")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		dimensions := parseDimensions(dimensionStr)
		lookback := request.GetInt("lookback_seconds", 3600)
		topx := request.GetInt("topx", 10)

		names, err := c.ListDeviceNames()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch devices: %v", err)), nil
		}

		query := map[string]any{
			"metric":           metric,
			"dimension":        dimensions,
			"topx":             topx,
			"depth":            100,
			"fastData":         "Auto",
			"outsort":          getOutsort(metric),
			"lookback_seconds": lookback,
			"time_format":      "UTC",
			"hostname_lookup":  true,
			"device_name":      strings.Join(names, ","),
			"all_selected":     false,
			"viz_type":         "stackedArea",
		}

		body := map[string]any{
			"queries": []map[string]any{
				{
					"query":       query,
					"bucket":      "Left +Y Axis",
					"bucketIndex": 0,
					"isOverlay":   false,
				},
			},
		}

		data, err := c.V5("POST", "/query/topXdata", body)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to query time series: %v", err)), nil
		}

		return mcp.NewToolResultText(summarizeTimeSeriesResults(data, metric)), nil
	}
}

func parseDimensions(s string) []string {
	var dims []string
	for _, d := range strings.Split(s, ",") {
		d = strings.TrimSpace(d)
		if d != "" {
			dims = append(dims, d)
		}
	}
	return dims
}

func getOutsort(metric string) string {
	switch metric {
	case "packets", "in_packets", "out_packets":
		return "avg_pkts_per_sec"
	case "fps":
		return "avg_flows_per_sec"
	case "unique_src_ip":
		return "avg_src_ip"
	case "unique_dst_ip":
		return "avg_dst_ip"
	default:
		return "avg_bits_per_sec"
	}
}

func summarizeFlowResults(data json.RawMessage, metric string) string {
	var resp struct {
		Results []struct {
			Data []map[string]any `json:"data"`
		} `json:"results"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return formatJSON(data)
	}

	if len(resp.Results) == 0 || len(resp.Results[0].Data) == 0 {
		return "No results returned.\n\n" + formatJSON(data)
	}

	rows := resp.Results[0].Data
	valueKey := getOutsort(metric)

	var total float64
	for _, entry := range rows {
		if v, ok := entry[valueKey].(float64); ok {
			total += v
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Flow Query Results (%d rows)\n\n", len(rows)))
	sb.WriteString("| Key | Avg Rate | %% Total |\n")
	sb.WriteString("|-----|----------|--------|\n")

	for _, entry := range rows {
		key := fmt.Sprintf("%v", entry["key"])
		v, _ := entry[valueKey].(float64)
		pct := 0.0
		if total > 0 {
			pct = v / total * 100
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %.1f%% |\n", key, formatRate(v, metric), pct))
	}

	sb.WriteString("\n<details><summary>Raw JSON</summary>\n\n```json\n")
	sb.WriteString(formatJSON(data))
	sb.WriteString("\n```\n</details>\n")

	return sb.String()
}

func summarizeTimeSeriesResults(data json.RawMessage, metric string) string {
	var resp struct {
		Results []struct {
			Data []map[string]any `json:"data"`
		} `json:"results"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return formatJSON(data)
	}

	if len(resp.Results) == 0 || len(resp.Results[0].Data) == 0 {
		return "No results returned.\n\n" + formatJSON(data)
	}

	rows := resp.Results[0].Data
	valueKey := getOutsort(metric)
	p95Key := "p95th_bits_per_sec"
	maxKey := "max_bits_per_sec"
	if strings.Contains(metric, "packets") || metric == "packets" {
		p95Key = "p95th_pkts_per_sec"
		maxKey = "max_pkts_per_sec"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Time Series Results (%d entries)\n\n", len(rows)))
	sb.WriteString("| Key | Avg | P95 | Max |\n")
	sb.WriteString("|-----|-----|-----|-----|\n")

	for _, entry := range rows {
		key := fmt.Sprintf("%v", entry["key"])
		avg, _ := entry[valueKey].(float64)
		p95, _ := entry[p95Key].(float64)
		max, _ := entry[maxKey].(float64)
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			key, formatRate(avg, metric), formatRate(p95, metric), formatRate(max, metric)))
	}

	sb.WriteString("\n<details><summary>Raw JSON</summary>\n\n```json\n")
	sb.WriteString(formatJSON(data))
	sb.WriteString("\n```\n</details>\n")

	return sb.String()
}
