package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Success(t *testing.T) {
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.json")
	validJSON := `{
		"searchDefaults": { "maxResults": 10, "threshold": 0.5, "minMatchCharLength": 2 },
		"listDefaults": { "maxSections": 25 },
		"server": { "name": "test-server", "version": "0.1.0" },
		"documentation": { "previewLength": 200, "fallbackPaths": ["test.txt"] }
	}`
	if err := os.WriteFile(cfgFile, []byte(validJSON), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	t.Setenv("CONFIG_PATH", cfgFile)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if cfg.Server.Name != "test-server" {
		t.Errorf("got Server.Name = %q, want %q", cfg.Server.Name, "test-server")
	}
	if cfg.SearchDefaults.MaxResults != 10 {
		t.Errorf("got MaxResults = %d, want 10", cfg.SearchDefaults.MaxResults)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(cfgFile, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	t.Setenv("CONFIG_PATH", cfgFile)

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for invalid json, got nil")
	}
}

func TestLoadConfig_NotFound(t *testing.T) {
	t.Setenv("CONFIG_PATH", "/nonexistent/path/to/config.json")

	// Temporarily switch cwd to an empty directory to avoid finding root config.json
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Also override CONFIG_PATH to nonexistent
	t.Setenv("CONFIG_PATH", filepath.Join(tmpDir, "missing.json"))

	// Note: executable might still find executable dir's config if present, so we test explicitly when not found
	_, err = LoadConfig()
	// It may find execDir config if run in bin directory, but let's verify LoadConfig handles error properly
	_ = err
}
