// Package config loads and exposes the application configuration from config.json.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SearchConfig holds defaults for the search engine.
type SearchConfig struct {
	MaxResults        int     `json:"maxResults"`
	Threshold         float64 `json:"threshold"`
	MinMatchCharLength int    `json:"minMatchCharLength"`
}

// ListConfig holds defaults for listing sections.
type ListConfig struct {
	MaxSections int `json:"maxSections"`
}

// ServerConfig holds the MCP server identity.
type ServerConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// DocConfig holds documentation loading settings.
type DocConfig struct {
	PreviewLength int      `json:"previewLength"`
	FallbackPaths []string `json:"fallbackPaths"`
}

// AppConfig is the top-level application configuration.
type AppConfig struct {
	SearchDefaults SearchConfig `json:"searchDefaults"`
	ListDefaults   ListConfig   `json:"listDefaults"`
	Server         ServerConfig `json:"server"`
	Documentation  DocConfig    `json:"documentation"`
}

// LoadConfig reads config.json from candidate locations:
// 1. CONFIG_PATH environment variable (if set)
// 2. Current working directory: ./config.json
// 3. Executable directory: ./config.json
// 4. Executable parent directory: ../config.json
func LoadConfig() (*AppConfig, error) {
	var candidates []string

	if env := os.Getenv("CONFIG_PATH"); env != "" {
		candidates = append(candidates, env)
	}

	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, "config.json"))
	}

	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		candidates = append(candidates,
			filepath.Join(execDir, "config.json"),
			filepath.Join(execDir, "..", "config.json"),
		)
	}

	var data []byte
	var lastErr error
	var resolvedPath string

	for _, p := range candidates {
		d, err := os.ReadFile(p)
		if err == nil {
			data = d
			resolvedPath = p
			break
		}
		lastErr = err
	}

	if data == nil {
		return nil, fmt.Errorf("read config file: %w (tried %v)", lastErr, candidates)
	}
	_ = resolvedPath

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}
