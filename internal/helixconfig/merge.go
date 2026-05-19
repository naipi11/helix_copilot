package helixconfig

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// MergeCopilot merges the helix-copilot language server configuration into a
// Helix languages.toml file without replacing existing language entries.
func MergeCopilot(path string) error {
	root := map[string]any{}
	if data, err := os.ReadFile(path); err == nil && len(bytes.TrimSpace(data)) > 0 {
		if err := toml.Unmarshal(data, &root); err != nil {
			return fmt.Errorf("parse existing languages.toml: %w", err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read existing languages.toml: %w", err)
	}

	mergeCopilotConfig(root)

	var buf bytes.Buffer
	buf.WriteString("# Generated/updated by helix-copilot. Existing language entries are preserved.\n")
	if err := toml.NewEncoder(&buf).Encode(root); err != nil {
		return fmt.Errorf("encode merged languages.toml: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write merged languages.toml: %w", err)
	}
	return nil
}

func mergeCopilotConfig(root map[string]any) {
	languageServer := ensureMap(root, "language-server")
	languageServer["copilot"] = map[string]any{
		"command": "helix-copilot",
		"args":    []string{"lsp"},
	}

	pylsp := ensureMapPath(languageServer, "pylsp", "config", "pylsp", "plugins")
	ensureMap(pylsp, "pycodestyle")["enabled"] = false
	ensureMap(pylsp, "pyflakes")["enabled"] = false
	ensureMap(pylsp, "flake8")["enabled"] = false

	languages := normalizeLanguageArray(root["language"])
	for _, wanted := range copilotLanguages() {
		idx := findLanguage(languages, wanted.Name)
		if idx >= 0 {
			servers := toStringSlice(languages[idx]["language-servers"])
			languages[idx]["language-servers"] = appendUnique(servers, "copilot")
			continue
		}
		entry := map[string]any{"name": wanted.Name}
		if len(wanted.Scope) > 0 {
			entry["scope"] = wanted.Scope
		}
		if len(wanted.FileTypes) > 0 {
			entry["file-types"] = wanted.FileTypes
		}
		entry["language-servers"] = appendUnique(wanted.LanguageServers, "copilot")
		languages = append(languages, entry)
	}
	root["language"] = languages
}

type languageTemplate struct {
	Name            string
	Scope           string
	FileTypes       []string
	LanguageServers []string
}

func copilotLanguages() []languageTemplate {
	return []languageTemplate{
		{Name: "python", Scope: "source.python", FileTypes: []string{"py", "pyi"}, LanguageServers: []string{"pylsp"}},
		{Name: "go", Scope: "source.go", FileTypes: []string{"go"}, LanguageServers: []string{"gopls"}},
		{Name: "rust", Scope: "source.rust", FileTypes: []string{"rs"}},
		{Name: "javascript", Scope: "source.js", FileTypes: []string{"js", "mjs", "cjs"}, LanguageServers: []string{"typescript-language-server"}},
		{Name: "typescript", Scope: "source.ts", FileTypes: []string{"ts", "mts", "cts"}, LanguageServers: []string{"typescript-language-server"}},
	}
}

func ensureMap(root map[string]any, key string) map[string]any {
	if existing, ok := root[key].(map[string]any); ok {
		return existing
	}
	m := map[string]any{}
	root[key] = m
	return m
}

func ensureMapPath(root map[string]any, keys ...string) map[string]any {
	cur := root
	for _, key := range keys {
		cur = ensureMap(cur, key)
	}
	return cur
}

func normalizeLanguageArray(value any) []map[string]any {
	var out []map[string]any
	switch typed := value.(type) {
	case []map[string]any:
		return typed
	case []any:
		for _, item := range typed {
			if m, ok := item.(map[string]any); ok {
				out = append(out, m)
			}
		}
	}
	return out
}

func findLanguage(languages []map[string]any, name string) int {
	for i, lang := range languages {
		if langName, ok := lang["name"].(string); ok && langName == name {
			return i
		}
	}
	return -1
}

func toStringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case string:
		return []string{typed}
	default:
		return nil
	}
}

func appendUnique(values []string, value string) []string {
	out := append([]string(nil), values...)
	for _, existing := range out {
		if existing == value {
			return out
		}
	}
	return append(out, value)
}
