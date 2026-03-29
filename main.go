package main

import (
	"fmt"
	"os"

	"github.com/hervehildenbrand/kentik-mcp/internal/client"
	"github.com/hervehildenbrand/kentik-mcp/internal/config"
	"github.com/hervehildenbrand/kentik-mcp/internal/tools"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n\n", err)
		fmt.Fprintln(os.Stderr, "Option 1 — Environment variables:")
		fmt.Fprintln(os.Stderr, "  export KENTIK_EMAIL=user@example.com")
		fmt.Fprintln(os.Stderr, "  export KENTIK_API_TOKEN=your-api-token")
		fmt.Fprintln(os.Stderr, "  export KENTIK_REGION=EU  # EU (default) or US")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Option 2 — Config file (~/.kentik-mcp.json):")
		fmt.Fprintln(os.Stderr, `  {"profiles":{"default":{"email":"...","api_token":"...","region":"EU"}}}`)
		fmt.Fprintln(os.Stderr, "  export KENTIK_PROFILE=profile-name  # optional, defaults to 'default'")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Environment variables always take precedence over config file.")
		os.Exit(1)
	}

	kentikClient := client.New(cfg)

	s := server.NewMCPServer(
		"kentik-mcp",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	tools.RegisterAll(s, kentikClient)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
