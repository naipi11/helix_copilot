//go:build !windows

package lsp

import (
	"os/exec"

	"github.com/naipi11/helix_copilot/internal/config"
	"github.com/naipi11/helix_copilot/internal/proxylog"
)

func buildChildCommand(cfg config.Config, log *proxylog.Logger) *exec.Cmd {
	pkg := cfg.LanguageServerPackage
	if pkg == "" {
		pkg = config.DefaultLanguageServerPackage
	}
	log.Printf("launcher: spawning npx --yes %s --stdio", pkg)
	return exec.Command("npx", "--yes", pkg, "--stdio")
}
