package k8s

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
		StructuredContent: output,
	}
}

// RegisterTools registers all Kubernetes MCP tools
func RegisterTools(server *mcp.Server, client *Client) {
	// Cluster info tools
	registerClusterTools(server, client)

	// Namespace and node tools
	registerNamespaceTools(server, client)

	// Pod tools
	registerPodTools(server, client)

	// Deployment tools
	registerDeploymentTools(server, client)

	// Service tools
	registerServiceTools(server, client)

	// Event tools
	registerEventTools(server, client)

	// Action tools (write operations)
	RegisterActionTools(server, client)
}
