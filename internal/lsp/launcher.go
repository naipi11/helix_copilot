package lsp

import (
	"os/exec"

	"github.com/naipi11/helix_copilot/internal/config"
	"github.com/naipi11/helix_copilot/internal/proxylog"
)

// BuildChildCommand constructs the *exec.Cmd that launches the Copilot
// language server. The implementation is platform-specific:
//
//   - On Linux/macOS it returns `npx --yes <pkg> --stdio`, identical to
//     the original behavior.
//   - On Windows it tries to locate an already-installed language server
//     entry .js and run it via node.exe directly, bypassing the npx.cmd
//     batch shim and its cmd.exe wrapper. Auto-installs the package on
//     first miss, falling back to npx.cmd as a last resort.
//
// The returned cmd has nothing wired up yet; the caller is responsible for
// stdin/stdout/stderr pipes.
func BuildChildCommand(cfg config.Config, log *proxylog.Logger) *exec.Cmd {
	return buildChildCommand(cfg, log)
}
