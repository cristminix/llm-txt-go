// Package mcp provides the MCP server setup using the mark3labs/mcp-go library.
package mcp

import (
	"github.com/mark3labs/mcp-go/server"

	appconfig "github.com/llm-txt-go-go/internal/config"
	"github.com/llm-txt-go-go/internal/tools"
)

// MedusaDocsServer wraps the MCP server and manages its lifecycle.
type MedusaDocsServer struct {
	config       *appconfig.AppConfig
	toolRegistry *tools.ToolRegistry
	mcpServer    *server.MCPServer
}

// NewMedusaDocsServer creates and configures a MedusaDocsServer.
func NewMedusaDocsServer(cfg *appconfig.AppConfig, tr *tools.ToolRegistry) *MedusaDocsServer {
	mcpServer := server.NewMCPServer(
		cfg.Server.Name,
		cfg.Server.Version,
		server.WithToolCapabilities(true),
	)

	// Register all tools with the MCP server
	for _, tool := range tr.ListTools() {
		handler := tr.Handler(tool.Name)
		if handler != nil {
			mcpServer.AddTool(tool, server.ToolHandlerFunc(handler))
		}
	}

	return &MedusaDocsServer{
		config:       cfg,
		toolRegistry: tr,
		mcpServer:    mcpServer,
	}
}

// Start begins serving the MCP server over stdio.
func (s *MedusaDocsServer) Start() error {
	return server.ServeStdio(s.mcpServer)
}
