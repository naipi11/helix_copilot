package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/naipi11/helix_copilot/internal/config"
)

func TestModelCommandPersistsSelection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	exitCode := run([]string{"model", "gpt-4o-copilot", "--config", path})
	if exitCode != 0 {
		t.Fatalf("run model exit = %d", exitCode)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Model != "gpt-4o-copilot" {
		t.Fatalf("model = %q", cfg.Model)
	}
}

func TestHelixLanguageSnippetContainsCopilotServer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "languages.toml")
	exitCode := run([]string{"configure-helix", "--output", path})
	if exitCode != 0 {
		t.Fatalf("configure-helix exit = %d", exitCode)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read snippet: %v", err)
	}
	text := string(data)
	for _, needle := range []string{
		"[language-server.copilot]",
		"helix-copilot",
		"lsp",
		"language-servers",
		"[language-server.pylsp.config.pylsp.plugins.pycodestyle]",
		"name = \"python\"",
		"language-servers = [\"pylsp\", \"copilot\"]",
	} {
		if !contains(text, needle) {
			t.Fatalf("snippet missing %q:\n%s", needle, text)
		}
	}
}

func contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}
