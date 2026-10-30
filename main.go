// llm-txt-go-go is an MCP server that provides local Medusa.js documentation
// via tools for AI assistants. It is a production-ready port of the Node.js version.
package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"

	"github.com/llm-txt-go-go/internal/config"
	"github.com/llm-txt-go-go/internal/docloader"
	mcpserver "github.com/llm-txt-go-go/internal/mcp"
	"github.com/llm-txt-go-go/internal/search"
	"github.com/llm-txt-go-go/internal/tools"
)

// loadEnv loads environment variables from a .env file if available.
func loadEnv() {
	if err := godotenv.Load(); err != nil {
		if execPath, execErr := os.Executable(); execErr == nil {
			_ = godotenv.Load(filepath.Join(filepath.Dir(execPath), ".env"))
		}
	}
}

// getDocName returns the documentation name from DOCUMENTATION_NAME env var,
// falling back to "Medusa Docs" if not set.
func getDocName() string {
	if name := os.Getenv("DOCUMENTATION_NAME"); name != "" {
		return name
	}
	return "Medusa Docs"
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	loadEnv()

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	docLoader := docloader.NewDocLoader(cfg)
	sections := docLoader.LoadSections()

	searchEngine := search.NewSearchEngine(sections, cfg.SearchDefaults)

	toolRegistry := tools.NewToolRegistry()

	searchTool, searchHandler := tools.CreateSearchDocsTool(searchEngine, cfg)
	toolRegistry.Register(searchTool, searchHandler)

	getSectionTool, getSectionHandler := tools.CreateGetSectionTool(sections)
	toolRegistry.Register(getSectionTool, getSectionHandler)

	listSectionsTool, listSectionsHandler := tools.CreateListSectionsTool(sections, cfg)
	toolRegistry.Register(listSectionsTool, listSectionsHandler)

	server := mcpserver.NewMedusaDocsServer(cfg, toolRegistry)

	docName := getDocName()
	slog.Info(fmt.Sprintf("%s MCP Server starting on stdio", docName))
	if err := server.Start(); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}
