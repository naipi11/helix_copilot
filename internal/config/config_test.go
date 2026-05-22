package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefaultConfigPathUsesXDGConfigHome(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses Unix paths")
	}
	t.Setenv("XDG_CONFIG_HOME", "/tmp/hxconf")
	t.Setenv("HOME", "/home/example")

	got := DefaultPath()
	want := filepath.Join("/tmp/hxconf", "helix-copilot", "config.json")
	if got != want {
		t.Fatalf("DefaultPath() = %q, want %q", got, want)
	}
}

func TestDefaultConfigPathFallsBackToHome(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses Unix paths")
	}
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/home/example")

	got := DefaultPath()
	want := filepath.Join("/home/example", ".config", "helix-copilot", "config.json")
	if got != want {
		t.Fatalf("DefaultPath() = %q, want %q", got, want)
	}
}

func TestDefaultConfigPathUsesAppDataOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-specific test")
	}
	t.Setenv("APPDATA", `C:\Users\test\AppData\Roaming`)

	got := DefaultPath()
	want := filepath.Join(`C:\Users\test\AppData\Roaming`, "helix-copilot", "config.json")
	if got != want {
		t.Fatalf("DefaultPath() = %q, want %q", got, want)
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load(missing) returned error: %v", err)
	}
	if cfg.Model != DefaultModel {
		t.Fatalf("Model = %q, want %q", cfg.Model, DefaultModel)
	}
	if cfg.LanguageServerPackage != DefaultLanguageServerPackage {
		t.Fatalf("LanguageServerPackage = %q, want %q", cfg.LanguageServerPackage, DefaultLanguageServerPackage)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")
	cfg := Config{Model: "gpt-4o-copilot", LanguageServerPackage: "@github/copilot-language-server"}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("config file not written: %v", err)
	}
	if runtime.GOOS != "windows" {
		if info.Mode().Perm() != 0o600 {
			t.Fatalf("mode = %v, want 0600", info.Mode().Perm())
		}
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got != cfg {
		t.Fatalf("Load() = %#v, want %#v", got, cfg)
	}
}
