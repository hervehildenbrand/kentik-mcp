package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerAIAdvisorTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(mcp.NewTool("ask_kentik_ai",
		mcp.WithDescription("Ask Kentik's AI Advisor a natural language question about your network. The AI autonomously queries your Kentik data, performs analysis, and returns a formatted answer with tables and Data Explorer links. "+
			"Responses take 15-60 seconds (async polling). "+
			"Best for: 'What caused the traffic spike at 3pm?', 'Which interfaces are near capacity?', "+
			"'Compare this week to last week', 'Show me anomalies in the last 24h', "+
			"'What are my top 5 traffic sources?', 'Is there unusual traffic to port 53?'. "+
			"For precise queries with specific dimensions/filters, use query_flows directly instead."),
		mcp.WithString("question", mcp.Required(),
			mcp.Description("Natural language question about your network data")),
	), makeAIAdvisorHandler(c))
}

func makeAIAdvisorHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		question, err := request.RequireString("question")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Create chat session
		body := map[string]any{
			"prompt": question,
		}
		data, err := c.V6("POST", "/ai_advisor/v202511/chat", body)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create AI session: %v", err)), nil
		}

		var session struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}
		if err := json.Unmarshal(data, &session); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to parse session: %v", err)), nil
		}

		if session.ID == "" {
			return mcp.NewToolResultError("AI Advisor returned empty session ID"), nil
		}

		// Poll for response (max 90 seconds)
		pollPath := fmt.Sprintf("/ai_advisor/v202511/chat/%s", session.ID)
		deadline := time.Now().Add(90 * time.Second)

		for time.Now().Before(deadline) {
			select {
			case <-ctx.Done():
				return mcp.NewToolResultError("Request cancelled"), nil
			case <-time.After(5 * time.Second):
			}

			pollData, err := c.V6("GET", pollPath, nil)
			if err != nil {
				continue
			}

			var resp struct {
				ID       string `json:"id"`
				Status   string `json:"status"`
				Messages []struct {
					Status      string `json:"status"`
					FinalAnswer string `json:"finalAnswer"`
					Reasoning   string `json:"reasoning"`
					Error       string `json:"errorMessage"`
				} `json:"messages"`
			}
			if err := json.Unmarshal(pollData, &resp); err != nil {
				continue
			}

			switch resp.Status {
			case "SESSION_STATUS_PENDING", "SESSION_STATUS_PROCESSING", "SESSION_STATUS_RUNNING":
				continue
			case "SESSION_STATUS_ERROR":
				errMsg := "AI Advisor returned an error"
				if len(resp.Messages) > 0 && resp.Messages[0].Error != "" {
					errMsg = resp.Messages[0].Error
				}
				return mcp.NewToolResultError(errMsg), nil
			default:
				// Completed
				if len(resp.Messages) > 0 && resp.Messages[0].FinalAnswer != "" {
					var sb strings.Builder
					sb.WriteString("## Kentik AI Advisor\n\n")
					sb.WriteString(fmt.Sprintf("**Question:** %s\n\n", question))
					sb.WriteString(resp.Messages[0].FinalAnswer)
					return mcp.NewToolResultText(sb.String()), nil
				}
				// No final answer yet but status changed
				if len(resp.Messages) > 0 && resp.Messages[0].Reasoning != "" {
					return mcp.NewToolResultText(fmt.Sprintf("AI Advisor is still thinking:\n\n%s", resp.Messages[0].Reasoning)), nil
				}
				return mcp.NewToolResultText(formatJSON(pollData)), nil
			}
		}

		return mcp.NewToolResultError("AI Advisor timed out after 90 seconds. The query may be too complex — try a simpler question."), nil
	}
}
