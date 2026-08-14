// Package tools provides MCP tools for the chile-bcn-mcp server.
package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"dev.alvaros.chile-bcn-mcp/internal/bcn"
)

// RegisterTools registers all tools on the MCP server. The law client is
// injected: the same process-wide instance is reused for every request.
func RegisterTools(srv *mcp.Server, client bcn.LawClient) {
	RegisterSearchLaws(srv, client)
	RegisterGetLaw(srv, client)
	RegisterGetLawSummary(srv, client)
}

// errorResult renders a tool error as an MCP error result. The handler
// still returns nil error: the failure is part of the tool response, so
// the client receives the message instead of a protocol error.
func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
	}
}
