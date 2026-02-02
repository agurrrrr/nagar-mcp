package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolError creates an MCP error result with the given message
func ToolError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
		IsError: true,
	}
}

// ToolErrorf creates an MCP error result with formatted message
func ToolErrorf(format string, args ...any) *mcp.CallToolResult {
	return ToolError(fmt.Sprintf(format, args...))
}

// ToolText creates an MCP text result
func ToolText(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
	}
}

// ToolTextf creates an MCP text result with formatted message
func ToolTextf(format string, args ...any) *mcp.CallToolResult {
	return ToolText(fmt.Sprintf(format, args...))
}

// ToolJSON creates an MCP result with JSON content
func ToolJSON(data any) (*mcp.CallToolResult, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return ToolError("failed to serialize response: " + err.Error()), nil
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}
