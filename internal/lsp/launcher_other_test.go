//go:build !windows

package lsp

import (
	"strings"
	"testing"

	"github.com/naipi11/helix_copilot/internal/config"
	"github.com/naipi11/helix_copilot/internal/proxylog"
)

func TestBuildChildCommandUsesNpxOnUnix(t *testing.T) {
	cfg := config.Defaults()
	cmd := BuildChildCommand(cfg, &proxylog.Logger{})
	if cmd == nil {
		t.Fatal("BuildChildCommand returned nil")
	}
	if !strings.HasSuffix(cmd.Args[0], "npx") {
		t.Fatalf("argv[0] = %q, want suffix 'npx'", cmd.Args[0])
	}
	want := []string{"--yes", config.DefaultLanguageServerPackage, "--stdio"}
	if len(cmd.Args) != 1+len(want) {
		t.Fatalf("argv = %v, want npx + %v", cmd.Args, want)
	}
	for i, w := range want {
		if cmd.Args[1+i] != w {
			t.Fatalf("argv[%d] = %q, want %q", 1+i, cmd.Args[1+i], w)
		}
	}
}

func TestBuildChildCommandHonorsCustomPackage(t *testing.T) {
	cfg := config.Defaults()
	cfg.LanguageServerPackage = "@example/custom-ls"
	cmd := BuildChildCommand(cfg, &proxylog.Logger{})
	if got := cmd.Args[2]; got != "@example/custom-ls" {
		t.Fatalf("package arg = %q, want @example/custom-ls", got)
	}
}
