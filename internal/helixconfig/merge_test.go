package helixconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergeCopilotPreservesExistingLanguageConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "languages.toml")
	original := `# user comment
[language-server.rust-analyzer]
command = "rust-analyzer"

[[language]]
name = "rust"
auto-format = true
language-servers = ["rust-analyzer"]

[language.debugger]
name = "lldb-dap"
command = "lldb-dap"

[[grammar]]
name = "rust"
source = { git = "https://github.com/tree-sitter/tree-sitter-rust", rev = "abc" }
`
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := MergeCopilot(path); err != nil {
		t.Fatalf("MergeCopilot: %v", err)
	}
	merged := readString(t, path)
	for _, want := range []string{
		"[language-server.rust-analyzer]",
		"command = \"rust-analyzer\"",
		"name = \"rust\"",
		"auto-format = true",
		"language-servers = [\"rust-analyzer\", \"copilot\"]",
		"[language.debugger]",
		"name = \"lldb-dap\"",
		"[[grammar]]",
		"[language-server.copilot]",
	} {
		if !strings.Contains(merged, want) {
			t.Fatalf("merged config missing %q:\n%s", want, merged)
		}
	}
}

func TestMergeCopilotIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "languages.toml")
	if err := os.WriteFile(path, []byte(`[[language]]
name = "python"
language-servers = ["pylsp"]
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := MergeCopilot(path); err != nil {
		t.Fatalf("first merge: %v", err)
	}
	first := readString(t, path)
	if err := MergeCopilot(path); err != nil {
		t.Fatalf("second merge: %v", err)
	}
	second := readString(t, path)
	if first != second {
		t.Fatalf("merge is not idempotent\nfirst:\n%s\nsecond:\n%s", first, second)
	}
	if strings.Count(second, "copilot") < 1 {
		t.Fatalf("expected copilot in merged config:\n%s", second)
	}
	if strings.Count(second, `language-servers = ["pylsp", "copilot"]`) != 1 {
		t.Fatalf("expected python copilot language server entry exactly once, got:\n%s", second)
	}
}

func TestMergeCopilotCreatesNewFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "languages.toml")
	if err := MergeCopilot(path); err != nil {
		t.Fatalf("MergeCopilot: %v", err)
	}
	merged := readString(t, path)
	for _, want := range []string{
		"[language-server.copilot]",
		"[[language]]",
		"name = \"python\"",
		"language-servers = [\"pylsp\", \"copilot\"]",
	} {
		if !strings.Contains(merged, want) {
			t.Fatalf("new config missing %q:\n%s", want, merged)
		}
	}
}

func readString(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
