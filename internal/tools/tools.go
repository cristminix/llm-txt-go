// Package tools provides the tool registry and all MCP tool implementations.
package tools

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/llm-txt-go-go/internal/config"
	"github.com/llm-txt-go-go/internal/docloader"
	"github.com/llm-txt-go-go/internal/search"
)

// textResponse creates a CallToolResult with a single text content block.
func textResponse(text string) *mcp.CallToolResult {
	return mcp.NewToolResultText(text)
}

// ── Tool factories ───────────────────────────────────────────────────────────

// CreateSearchDocsTool returns the search_docs tool definition and handler.
func CreateSearchDocsTool(engine *search.SearchEngine, cfg *config.AppConfig) (mcp.Tool, ToolHandler) {
	docName := os.Getenv("DOCUMENTATION_NAME")
	if docName == "" {
		docName = "Medusa.js"
	}

	tool := mcp.NewTool("search_docs",
		mcp.WithDescription(fmt.Sprintf("Search %s documentation for specific topics", docName)),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query for finding relevant documentation"),
		),
		mcp.WithNumber("limit",
			mcp.Description(fmt.Sprintf("Maximum number of results to return (default: %d)", cfg.SearchDefaults.MaxResults)),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.Params.Arguments

		query, _ := args["query"].(string)
		limit := cfg.SearchDefaults.MaxResults
		if l, ok := args["limit"].(float64); ok && l > 0 {
			limit = int(l)
		}

		results := engine.Search(query, limit)

		if len(results) == 0 {
			return textResponse(fmt.Sprintf("No results found for query: \"%s\"", query)), nil
		}

		var formatted []string
		for i, r := range results {
			preview := r.Item.Content
			if len(preview) > cfg.Documentation.PreviewLength {
				preview = preview[:cfg.Documentation.PreviewLength] + "..."
			}
			formatted = append(formatted, fmt.Sprintf("%d. **%s**\n%s\n", i+1, r.Item.Title, preview))
		}

		return textResponse(strings.Join(formatted, "\n---\n")), nil
	}

	return tool, handler
}

// CreateGetSectionTool returns the get_section tool definition and handler.
func CreateGetSectionTool(sections []docloader.DocSection) (mcp.Tool, ToolHandler) {
	tool := mcp.NewTool("get_section",
		mcp.WithDescription("Get a specific documentation section by title or path"),
		mcp.WithString("identifier",
			mcp.Required(),
			mcp.Description("Section title or path to retrieve"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		identifier, _ := request.Params.Arguments["identifier"].(string)
		identifierLower := strings.ToLower(identifier)

		var section *docloader.DocSection
		for i := range sections {
			s := &sections[i]
			if strings.Contains(strings.ToLower(s.Title), identifierLower) ||
				(s.Path != "" && strings.Contains(strings.ToLower(s.Path), identifierLower)) {
				section = s
				break
			}
		}

		if section == nil {
			return textResponse(fmt.Sprintf("Section not found: \"%s\"", identifier)), nil
		}

		return textResponse(fmt.Sprintf("# %s\n\n%s", section.Title, section.Content)), nil
	}

	return tool, handler
}

// CreateListSectionsTool returns the list_sections tool definition and handler.
func CreateListSectionsTool(sections []docloader.DocSection, cfg *config.AppConfig) (mcp.Tool, ToolHandler) {
	tool := mcp.NewTool("list_sections",
		mcp.WithDescription("List all available documentation sections"),
		mcp.WithString("category",
			mcp.Description("Filter sections by category (optional)"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		category, _ := request.Params.Arguments["category"].(string)
		categoryLower := strings.ToLower(category)

		filtered := sections
		if categoryLower != "" {
			var matched []docloader.DocSection
			for _, s := range sections {
				if strings.Contains(strings.ToLower(s.Path), categoryLower) ||
					strings.Contains(strings.ToLower(s.Title), categoryLower) {
					matched = append(matched, s)
				}
			}
			filtered = matched
		}

		maxSections := cfg.ListDefaults.MaxSections
		if len(filtered) > maxSections {
			filtered = filtered[:maxSections]
		}

		var list []string
		for _, s := range filtered {
			list = append(list, fmt.Sprintf("- %s", s.Title))
		}

		header := "Available sections:"
		if categoryLower != "" {
			header = fmt.Sprintf("Available sections (filtered by: %s):", category)
		}

		return textResponse(fmt.Sprintf("%s\n\n%s\n\nTotal: %d sections", header, strings.Join(list, "\n"), len(filtered))), nil
	}

	return tool, handler
}
