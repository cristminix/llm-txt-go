// Package tools provides the tool registry and all MCP tool implementations.
package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// ToolHandler is the function signature for handling a tool call.
type ToolHandler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

// ToolRegistry manages registration and execution of MCP tools.
type ToolRegistry struct {
	tools     []mcp.Tool
	handlers  map[string]ToolHandler
}

// NewToolRegistry creates an empty ToolRegistry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		handlers: make(map[string]ToolHandler),
	}
}

// Register adds a tool definition and its handler to the registry.
func (tr *ToolRegistry) Register(tool mcp.Tool, handler ToolHandler) {
	tr.tools = append(tr.tools, tool)
	tr.handlers[tool.Name] = handler
}

// ListTools returns all registered tool definitions.
func (tr *ToolRegistry) ListTools() []mcp.Tool {
	return tr.tools
}

// Handler returns the ToolHandler for the named tool, or nil if not found.
func (tr *ToolRegistry) Handler(name string) ToolHandler {
	return tr.handlers[name]
}

// Execute runs the handler for the named tool with the given arguments.
func (tr *ToolRegistry) Execute(ctx context.Context, name string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	handler, ok := tr.handlers[name]
	if !ok {
		return mcp.NewToolResultText(fmt.Sprintf("Unknown tool: %s", name)), nil
	}

	var request mcp.CallToolRequest
	request.Params.Name = name
	request.Params.Arguments = args

	return handler(ctx, request)
}
