package tools

import (
	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterAll(s *server.MCPServer, c *client.Client) {
	registerQueryTool(s, c)
	registerInventoryTool(s, c)
	registerAlertsTool(s, c)
	registerSyntheticsTool(s, c)
	registerConfigTool(s, c)
	registerBGPTool(s, c)
	registerAITool(s, c)
	registerMarketTool(s, c)
}
