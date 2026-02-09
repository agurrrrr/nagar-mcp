package opnsense

import (
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Helper functions for typed results ===

func toolErrorFor[Out any](msg string) *mcp.CallToolResultFor[Out] {
	return &mcp.CallToolResultFor[Out]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
		IsError: true,
	}
}

func toolErrorfFor[Out any](format string, args ...any) *mcp.CallToolResultFor[Out] {
	return toolErrorFor[Out](fmt.Sprintf(format, args...))
}

func toolSuccessFor[Out any](output Out) *mcp.CallToolResultFor[Out] {
	return &mcp.CallToolResultFor[Out]{
		Content:           []mcp.Content{},
		StructuredContent: output,
	}
}

// RegisterTools registers all OPNsense MCP tools
func RegisterTools(server *mcp.Server, client *Client) {
	registerSystemTools(server, client)
	registerFirewallTools(server, client)
	registerDHCPTools(server, client)
	registerDNSTools(server, client)
	registerServiceTools(server, client)
	registerWireGuardTools(server, client)
	registerHAProxyTools(server, client)
}
