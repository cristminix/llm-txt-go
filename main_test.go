package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDocName_Default(t *testing.T) {
	// Ensure DOCUMENTATION_NAME is not set
	orig := os.Getenv("DOCUMENTATION_NAME")
	defer os.Setenv("DOCUMENTATION_NAME", orig)

	os.Unsetenv("DOCUMENTATION_NAME")
	if got := getDocName(); got != "Medusa Docs" {
		t.Errorf("getDocName() = %q, want %q", got, "Medusa Docs")
	}

	os.Setenv("DOCUMENTATION_NAME", "")
	if got := getDocName(); got != "Medusa Docs" {
		t.Errorf("getDocName() when empty = %q, want %q", got, "Medusa Docs")
	}
}

func TestGetDocName_Custom(t *testing.T) {
	orig := os.Getenv("DOCUMENTATION_NAME")
	defer os.Setenv("DOCUMENTATION_NAME", orig)

	os.Setenv("DOCUMENTATION_NAME", "React Native Docs")
	if got := getDocName(); got != "React Native Docs" {
		t.Errorf("getDocName() = %q, want %q", got, "React Native Docs")
	}
}

func TestLoadEnv(t *testing.T) {
	// Create temporary directory with a .env file
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte("DOCUMENTATION_NAME=\"Custom Loaded Docs\"\n"), 0644); err != nil {
		t.Fatalf("failed to write test .env: %v", err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer os.Chdir(origDir)

	origDocName := os.Getenv("DOCUMENTATION_NAME")
	defer os.Setenv("DOCUMENTATION_NAME", origDocName)

	os.Unsetenv("DOCUMENTATION_NAME")

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir to tmpDir: %v", err)
	}

	loadEnv()

	if got := getDocName(); got != "Custom Loaded Docs" {
		t.Errorf("after loadEnv(), getDocName() = %q, want %q", got, "Custom Loaded Docs")
	}
}
