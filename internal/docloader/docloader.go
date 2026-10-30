// Package docloader reads and parses the Medusa.js documentation file into sections.
package docloader

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/llm-txt-go-go/internal/config"
)

// DocSection represents a parsed section of documentation.
type DocSection struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	URL     string `json:"url,omitempty"`
	Path    string `json:"path,omitempty"`
}

// DocLoader reads and parses documentation from a markdown file.
type DocLoader struct {
	sections []DocSection
}

// NewDocLoader creates a DocLoader, resolves the documentation path, and parses sections.
func NewDocLoader(cfg *config.AppConfig) *DocLoader {
	dl := &DocLoader{}
	docsPath := dl.resolvePath(cfg)

	if docsPath == "" {
		tried := []string{os.Getenv("LLM_DOCS_PATH")}
		tried = append(tried, cfg.Documentation.FallbackPaths...)
		var filtered []string
		for _, p := range tried {
			if p != "" {
				filtered = append(filtered, p)
			}
		}
		slog.Error("Documentation file not found. Tried paths:", "paths", filtered)
		slog.Error("Please set LLM_DOCS_PATH environment variable or place llms-full.txt in the project directory")
		return dl
	}

	slog.Info(fmt.Sprintf("Loading documentation from: %s", docsPath))
	raw, err := os.ReadFile(docsPath)
	if err != nil {
		slog.Error("Failed to read documentation file", "path", docsPath, "error", err)
		return dl
	}

	dl.sections = dl.parse(string(raw))
	slog.Info(fmt.Sprintf("Successfully loaded documentation: %d sections parsed", len(dl.sections)))
	return dl
}

// LoadSections returns the parsed documentation sections.
func (dl *DocLoader) LoadSections() []DocSection {
	return dl.sections
}

// resolvePath finds the documentation file from env var or fallback paths.
func (dl *DocLoader) resolvePath(cfg *config.AppConfig) string {
	var possiblePaths []string

	if envPath := os.Getenv("LLM_DOCS_PATH"); envPath != "" {
		possiblePaths = append(possiblePaths, envPath)
	}

	cwd, _ := os.Getwd()
	for _, p := range cfg.Documentation.FallbackPaths {
		if strings.HasPrefix(p, "/") {
			possiblePaths = append(possiblePaths, p)
		} else {
			possiblePaths = append(possiblePaths, filepath.Join(cwd, p))
		}
	}

	for _, p := range possiblePaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
}

// parse splits raw markdown into DocSections based on H1 and H2 headers.
func (dl *DocLoader) parse(raw string) []DocSection {
	lines := strings.Split(raw, "\n")
	var sections []DocSection
	var currentSection *DocSection
	var contentBuffer []string

	for _, line := range lines {
		isH1 := strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "## ")
		isH2 := strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ")

		if isH1 {
			if currentSection != nil && len(contentBuffer) > 0 {
				currentSection.Content = strings.TrimSpace(strings.Join(contentBuffer, "\n"))
				sections = append(sections, *currentSection)
			}

			currentSection = &DocSection{
				Title: strings.TrimSpace(strings.TrimPrefix(line, "# ")),
				Path:  strings.ToLower(strings.ReplaceAll(strings.TrimSpace(strings.TrimPrefix(line, "# ")), " ", "_")),
			}
			contentBuffer = nil
		} else if isH2 && currentSection != nil {
			if len(contentBuffer) > 0 {
				subSection := DocSection{
					Title:   fmt.Sprintf("%s | %s", currentSection.Title, strings.TrimSpace(strings.TrimPrefix(line, "## "))),
					Content: strings.TrimSpace(strings.Join(contentBuffer, "\n")),
					Path:    currentSection.Path,
				}
				if subSection.Content != "" {
					sections = append(sections, subSection)
				}
			}
			contentBuffer = []string{line}
		} else if currentSection != nil {
			contentBuffer = append(contentBuffer, line)
		}
	}

	if currentSection != nil && len(contentBuffer) > 0 {
		currentSection.Content = strings.TrimSpace(strings.Join(contentBuffer, "\n"))
		sections = append(sections, *currentSection)
	}

	return sections
}
