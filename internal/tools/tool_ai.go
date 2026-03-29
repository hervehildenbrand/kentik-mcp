package tools

import (
	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerAITool(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("kentik_ai",
		mcp.WithDescription("Ask Kentik AI Advisor a natural language question about your network. "+
			"The AI autonomously queries your data and returns formatted analysis with tables and portal links. "+
			"Takes 15-60 seconds. Best for: 'What caused the traffic spike?', 'Which interfaces are near capacity?', "+
			"'Compare this week to last week', 'Show anomalies in the last 24h'."),
		mcp.WithString("question", mcp.Required(),
			mcp.Description("Natural language question about your network")),
	), makeAIAdvisorHandler(c))
}
