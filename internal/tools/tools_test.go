package tools

import (
	"os"
	"testing"

	"github.com/llm-txt-go-go/internal/config"
	"github.com/llm-txt-go-go/internal/docloader"
	"github.com/llm-txt-go-go/internal/search"
)

func TestCreateSearchDocsTool_Description(t *testing.T) {
	cfg := &config.AppConfig{
		SearchDefaults: config.SearchConfig{
			MaxResults: 5,
		},
	}
	engine := search.NewSearchEngine([]docloader.DocSection{}, cfg.SearchDefaults)

	// Test default when DOCUMENTATION_NAME is unset
	orig := os.Getenv("DOCUMENTATION_NAME")
	defer os.Setenv("DOCUMENTATION_NAME", orig)

	os.Unsetenv("DOCUMENTATION_NAME")
	toolDefault, _ := CreateSearchDocsTool(engine, cfg)
	expectedDefault := "Search Medusa.js documentation for specific topics"
	if toolDefault.Description != expectedDefault {
		t.Errorf("tool.Description = %q, want %q", toolDefault.Description, expectedDefault)
	}

	// Test custom DOCUMENTATION_NAME
	os.Setenv("DOCUMENTATION_NAME", "React Native")
	toolCustom, _ := CreateSearchDocsTool(engine, cfg)
	expectedCustom := "Search React Native documentation for specific topics"
	if toolCustom.Description != expectedCustom {
		t.Errorf("tool.Description = %q, want %q", toolCustom.Description, expectedCustom)
	}
}
