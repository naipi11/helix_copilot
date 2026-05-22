//go:build windows

package lsp

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/naipi11/helix_copilot/internal/config"
	"github.com/naipi11/helix_copilot/internal/proxylog"
)

// installTimeout caps the synchronous npm install run on first miss.
// 60s is enough for a warm npm cache; cold cache may need more but at
// that point the user almost certainly wants to abort and install manually.
const installTimeout = 60 * time.Second

// lsLookup encapsulates the side-effecting bits of locator logic so tests
// can inject deterministic stubs. Production code wires these to real OS
// primitives via newDefaultLookup.
type lsLookup struct {
	statFn func(string) (os.FileInfo, error)
	envFn  func(string) string
	runCmd func(name string, args ...string) (string, error)
}

func newDefaultLookup() *lsLookup {
	return &lsLookup{
		statFn: os.Stat,
		envFn:  os.Getenv,
		runCmd: func(name string, args ...string) (string, error) {
			out, err := exec.Command(name, args...).Output()
			return string(out), err
		},
	}
}

// Find returns (absolute path to language-server.js, human-readable reason).
// An empty path means "not found" — caller should consider auto-install.
// The reason string is logged verbatim so a log file is self-explaining.
func (l *lsLookup) Find(cfg config.Config) (string, string) {
	if v := strings.TrimSpace(l.envFn("HELIX_COPILOT_LS_PATH")); v != "" {
		return v, "env override (HELIX_COPILOT_LS_PATH)"
	}
	if v := strings.TrimSpace(cfg.LanguageServerJSPath); v != "" {
		return v, "config override (languageServerJSPath)"
	}
	rel := filepath.Join("node_modules", "@github", "copilot-language-server", "dist", "language-server.js")
	if appData := l.envFn("APPDATA"); appData != "" {
		candidate := filepath.Join(appData, "npm", rel)
		if _, err := l.statFn(candidate); err == nil {
			return candidate, "found at %APPDATA%\\npm"
		}
	}
	if out, err := l.runCmd("npm.cmd", "root", "-g"); err == nil {
		root := strings.TrimSpace(out)
		if root != "" {
			candidate := filepath.Join(root, "@github", "copilot-language-server", "dist", "language-server.js")
			if _, err := l.statFn(candidate); err == nil {
				return candidate, "found via npm root -g"
			}
		}
	}
	return "", "not found"
}

func buildChildCommand(cfg config.Config, log *proxylog.Logger) *exec.Cmd {
	return newDefaultLookup().build(cfg, log)
}

// build performs the full Windows launch flow on a *lsLookup so it can be
// driven by tests with stubbed lookup behavior.
func (l *lsLookup) build(cfg config.Config, log *proxylog.Logger) *exec.Cmd {
	pkg := cfg.LanguageServerPackage
	if pkg == "" {
		pkg = config.DefaultLanguageServerPackage
	}

	js, reason := l.Find(cfg)
	log.Printf("launcher: LS lookup -> %s | path=%q", reason, js)

	if js == "" {
		log.Printf("launcher: attempting auto-install: npm.cmd install -g %s (timeout %s)", pkg, installTimeout)
		if out, err := l.installGlobal(pkg); err != nil {
			log.Printf("launcher: auto-install failed: %v | output:\n%s", err, out)
		} else {
			log.Printf("launcher: auto-install succeeded | output:\n%s", out)
			js, reason = l.Find(cfg)
			log.Printf("launcher: post-install LS lookup -> %s | path=%q", reason, js)
		}
	}

	if js != "" {
		nodePath := l.resolveNode()
		log.Printf("launcher: spawning %q %q --stdio", nodePath, js)
		return exec.Command(nodePath, js, "--stdio")
	}

	npxPath := l.resolveNpx()
	log.Printf("launcher: WARNING falling back to npx shim at %q (cmd.exe wrapper, may break LSP framing)", npxPath)
	log.Printf("launcher: to fix permanently run: npm install -g %s", pkg)
	return exec.Command(npxPath, "--yes", pkg, "--stdio")
}

// installGlobal runs `npm.cmd install -g <pkg>` with a hard timeout and
// returns its combined output. Output is treated as opaque diagnostic text,
// never as LSP frames, so cmd.exe shim quirks here don't break the proxy.
func (l *lsLookup) installGlobal(pkg string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), installTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "npm.cmd", "install", "-g", pkg).CombinedOutput()
	return string(out), err
}

// resolveNode returns an explicit absolute path to node.exe when possible,
// falling back to the bare name (which Go will resolve via PATH).
func (l *lsLookup) resolveNode() string {
	if p, err := exec.LookPath("node.exe"); err == nil {
		return p
	}
	if p, err := exec.LookPath("node"); err == nil {
		return p
	}
	return "node.exe"
}

// resolveNpx prefers npx.cmd explicitly so Go always invokes the .cmd shim
// via the same path the user observes from `where npx`. This makes log
// lines unambiguous when triaging the broken fallback path.
func (l *lsLookup) resolveNpx() string {
	if p, err := exec.LookPath("npx.cmd"); err == nil {
		return p
	}
	if p, err := exec.LookPath("npx"); err == nil {
		return p
	}
	return "npx.cmd"
}
