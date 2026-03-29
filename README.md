# kentik-mcp

A Model Context Protocol (MCP) server for the [Kentik](https://www.kentik.com) network observability platform. Provides 8 consolidated tools covering flow queries, device inventory, DDoS alerting, BGP routing, synthetic monitoring, AI-powered analysis, and more — all accessible from Claude Code, Claude Desktop, or any MCP-compatible client.

## Installation

### Prerequisites

- [Go](https://go.dev/dl/) 1.21 or later
- A [Kentik](https://www.kentik.com) account with API access

### Build from Source

```bash
git clone https://github.com/hervehildenbrand/kentik-mcp.git
cd kentik-mcp
go build -o kentik-mcp .
```

### Get Your Kentik API Token

1. Log in to [portal.kentik.com](https://portal.kentik.com) (US) or [portal.kentik.eu](https://portal.kentik.eu) (EU)
2. Go to **User Profile** (top-right menu) > **API Access**
3. Copy your **API Token** (or create one if you don't have it)
4. Note your **account email** — you'll need both

### Install on Claude Desktop

1. Build the binary (see above)

2. Open your Claude Desktop config file:
   - **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
   - **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

3. Add the `kentik` server to the `mcpServers` section:

```json
{
  "mcpServers": {
    "kentik": {
      "command": "/absolute/path/to/kentik-mcp",
      "env": {
        "KENTIK_EMAIL": "your-email@company.com",
        "KENTIK_API_TOKEN": "your-api-token-here",
        "KENTIK_REGION": "EU"
      }
    }
  }
}
```

4. **Restart Claude Desktop** — you should see a hammer icon with 8 Kentik tools

5. Try asking: *"What are my top traffic sources right now?"* or *"Show me any active DDoS alarms"*

### Install on Claude Code

Add to your project's `.mcp.json`:

```json
{
  "mcpServers": {
    "kentik": {
      "command": "/absolute/path/to/kentik-mcp",
      "env": {
        "KENTIK_EMAIL": "your-email@company.com",
        "KENTIK_API_TOKEN": "your-api-token-here",
        "KENTIK_REGION": "EU"
      }
    }
  }
}
```

### Quick Test

```bash
# Verify it works (should list 8 tools)
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
  KENTIK_EMAIL=your-email KENTIK_API_TOKEN=your-token KENTIK_REGION=EU \
  ./kentik-mcp
```

## Configuration

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `KENTIK_EMAIL` | Yes* | Kentik account email |
| `KENTIK_API_TOKEN` | Yes* | API token from Kentik portal (User Profile > API) |
| `KENTIK_REGION` | No | `EU` (default) or `US` — selects API endpoints |
| `KENTIK_PROFILE` | No | Named profile from config file (default: `default`) |
| `KENTIK_V5_URL` | No | Override V5 REST API base URL |
| `KENTIK_V6_URL` | No | Override V6 gRPC-gateway base URL |

*Required unless provided via config file.

### Config File (~/.kentik-mcp.json)

For multi-user or multi-region setups, create `~/.kentik-mcp.json`:

```json
{
  "profiles": {
    "default": {
      "email": "user@company.com",
      "api_token": "your-eu-token",
      "region": "EU"
    },
    "us-prod": {
      "email": "user@company.com",
      "api_token": "your-us-token",
      "region": "US"
    }
  }
}
```

Switch profiles: `export KENTIK_PROFILE=us-prod`

Environment variables always take precedence over config file values.

### Regional Endpoints

| Region | V5 REST API | V6 gRPC Gateway |
|--------|-------------|-----------------|
| EU | `https://api.kentik.eu/api/v5` | `https://grpc.api.kentik.eu` |
| US | `https://api.kentik.com/api/v5` | `https://grpc.api.kentik.com` |

## Tools (8)

Each tool uses an `action` parameter to select the specific operation.

| Tool | Actions | Description |
|------|---------|-------------|
| `kentik_query` | `flows`, `time_series`, `top_ddos` | Flow data analysis — top-X by dimension, time trends, DDoS ranking |
| `kentik_inventory` | `list_devices`, `get_device`, `list_interfaces`, `get_interface`, `list_sites`, `get_site`, `list_labels`, `get_label` | Network infrastructure inventory |
| `kentik_alerts` | `list_policies`, `get_policy`, `active_alarms`, `alert_history`, `top_attacks` | DDoS protection and alerting |
| `kentik_synthetics` | `list_tests`, `get_test`, `get_results`, `list_agents` | Synthetic monitoring tests and results |
| `kentik_config` | `list_saved_filters`, `get_saved_filter`, `list_custom_dimensions`, `get_custom_dimension`, `list_users`, `get_user`, `list_cloud_exports`, `get_cloud_export`, `list_notification_channels`, `get_notification_channel` | Configuration and admin objects |
| `kentik_bgp` | `list_monitors`, `get_routes` | BGP routing data and prefix monitoring |
| `kentik_ai` | *(direct question)* | Natural language queries via Kentik AI Advisor (15-60s) |
| `kentik_market` | `as_info`, `as_relationships` | AS number lookup and network relationships (KMI) |

### Example Usage

```
kentik_query(action="flows", metric="bytes", dimension="AS_dst", topx=10)
kentik_query(action="time_series", metric="bytes", dimension="i_device_site_name", lookback_seconds=604800)
kentik_inventory(action="list_devices")
kentik_inventory(action="get_device", device_id="12345")
kentik_alerts(action="top_attacks", sort_by="bps", top=5)
kentik_alerts(action="active_alarms")
kentik_synthetics(action="get_results", test_ids="70933", start_time="2026-03-28T12:00:00Z", end_time="2026-03-28T13:00:00Z")
kentik_ai(question="Which interfaces are near capacity?")
```

## Flow Query Reference

### Dimensions (group-by fields for `kentik_query` action=flows)

| Dimension | Description |
|-----------|-------------|
| `AS_src`, `AS_dst` | Source/destination Autonomous System |
| `IP_src`, `IP_dst` | Source/destination IP address |
| `Port_src`, `Port_dst` | Source/destination port |
| `Proto` | IP protocol (TCP, UDP, ICMP, ESP, etc.) |
| `tcp_flags` | TCP flags (ACK, SYN, PSH+ACK, RST, etc.) |
| `Geography_src`, `Geography_dst` | Source/destination country |
| `InterfaceID_src`, `InterfaceID_dst` | Ingress/egress interface |
| `i_device_site_name` | Device site name |
| `i_device_name`, `i_device_id` | Device name or ID |
| `IP_dst_cidr_24_64` | Destination /24 prefix |
| `TopFlow` | Individual flow records (proto-src:port->dst:port) |

Multiple dimensions can be combined: `"AS_src,Port_dst"` or `"tcp_flags,Port_dst"`.

### Metrics

| Metric | Description |
|--------|-------------|
| `bytes` | Total bytes (returns avg_bits_per_sec) |
| `in_bytes`, `out_bytes` | Directional bytes |
| `packets` | Total packets (returns avg_pkts_per_sec) |
| `in_packets`, `out_packets` | Directional packets |
| `fps` | Flows per second |
| `unique_src_ip`, `unique_dst_ip` | Unique IP count |

### Filter Fields (for `filters_json`)

| Field | Description | Example Value |
|-------|-------------|---------------|
| `dst_as`, `src_as` | Destination/source AS number | `"15169"` |
| `inet_dst_addr`, `inet_src_addr` | Destination/source IP or CIDR | `"198.51.100.0/24"` |
| `l4_dst_port`, `l4_src_port` | Destination/source port | `"443"` |
| `protocol` | IP protocol number | `"6"` (TCP) |
| `tcp_flags` | TCP flag bitmask | `"2"` (SYN) |
| `i_device_name` | Device name | `"edge-router-1"` |
| `i_src_network_bndry_name` | Network boundary | `"external"` |
| `i_output_snmp_alias` | Output interface description | `"INT:IX:FRANCE-IX:100G"` |
| `i_trf_profile` | Traffic profile | `"from outside, terminated inside"` |

Filter operators: `=`, `!=`, `contains`, `not_contains`.

## Rate Limits

| API Type | Concurrent | Per Minute | Per Hour |
|----------|-----------|------------|----------|
| Non-query (devices, BGP, etc.) | 1 | 20-60 | 3,750 |
| Query (flow data) | 4 | 30-100 | 1,500 |
| AI Advisor | — | 4 create | — |

## Contributing

```bash
# Clone and build
git clone https://github.com/hervehildenbrand/kentik-mcp.git
cd kentik-mcp
go build -o kentik-mcp .

# Run tests (no Kentik account needed — all tests use mocks)
go test ./... -v
```

Pull requests welcome. Please include tests for new tools.

## License

MIT — see [LICENSE](LICENSE)
