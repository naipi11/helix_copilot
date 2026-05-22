//go:build windows

package lsp

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/naipi11/helix_copilot/internal/config"
	"github.com/naipi11/helix_copilot/internal/proxylog"
)

// fakeFileInfo is a trivial os.FileInfo implementation for tests that only
// need the existence semantics of os.Stat — none of these methods are read.
type fakeFileInfo struct{}

func (fakeFileInfo) Name() string       { return "language-server.js" }
func (fakeFileInfo) Size() int64        { return 0 }
func (fakeFileInfo) Mode() os.FileMode  { return 0 }
func (fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (fakeFileInfo) IsDir() bool        { return false }
func (fakeFileInfo) Sys() any           { return nil }

// existsAt builds a stub statFn that returns success only for the given
// absolute paths and ENOENT for everything else.
func existsAt(paths ...string) func(string) (os.FileInfo, error) {
	set := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		set[p] = struct{}{}
	}
	return func(p string) (os.FileInfo, error) {
		if _, ok := set[p]; ok {
			return fakeFileInfo{}, nil
		}
		return nil, os.ErrNotExist
	}
}

// envFromMap returns a function compatible with os.Getenv backed by a map.
func envFromMap(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestFindEnvOverrideWinsOverEverything(t *testing.T) {
	l := &lsLookup{
		statFn: existsAt(),
		envFn:  envFromMap(map[string]string{"HELIX_COPILOT_LS_PATH": `D:\custom\language-server.js`}),
		runCmd: func(string, ...string) (string, error) { return "", errors.New("should not run") },
	}
	cfg := config.Defaults()
	cfg.LanguageServerJSPath = `Z:\ignored.js`
	got, reason := l.Find(cfg)
	if got != `D:\custom\language-server.js` {
		t.Fatalf("path = %q, want env override path", got)
	}
	if !strings.Contains(reason, "env override") {
		t.Fatalf("reason = %q, want env override", reason)
	}
}

func TestFindConfigOverrideWinsOverFilesystem(t *testing.T) {
	appData := `C:\Users\test\AppData\Roaming`
	appDataHit := appData + `\npm\node_modules\@github\copilot-language-server\dist\language-server.js`
	l := &lsLookup{
		statFn: existsAt(appDataHit),
		envFn:  envFromMap(map[string]string{"APPDATA": appData}),
		runCmd: func(string, ...string) (string, error) { return "", errors.New("should not run") },
	}
	cfg := config.Defaults()
	cfg.LanguageServerJSPath = `D:\pinned\language-server.js`
	got, reason := l.Find(cfg)
	if got != `D:\pinned\language-server.js` {
		t.Fatalf("path = %q, want config override path", got)
	}
	if !strings.Contains(reason, "config override") {
		t.Fatalf("reason = %q, want config override", reason)
	}
}

func TestFindAppDataHit(t *testing.T) {
	appData := `C:\Users\test\AppData\Roaming`
	hit := appData + `\npm\node_modules\@github\copilot-language-server\dist\language-server.js`
	l := &lsLookup{
		statFn: existsAt(hit),
		envFn:  envFromMap(map[string]string{"APPDATA": appData}),
		runCmd: func(string, ...string) (string, error) { return "", errors.New("should not run") },
	}
	got, reason := l.Find(config.Defaults())
	if got != hit {
		t.Fatalf("path = %q, want %q", got, hit)
	}
	if !strings.Contains(reason, "APPDATA") {
		t.Fatalf("reason = %q, want APPDATA mention", reason)
	}
}

func TestFindNpmRootFallback(t *testing.T) {
	root := `C:\global\npm`
	hit := root + `\@github\copilot-language-server\dist\language-server.js`
	l := &lsLookup{
		statFn: existsAt(hit),
		envFn:  envFromMap(map[string]string{}), // APPDATA empty
		runCmd: func(name string, args ...string) (string, error) {
			if name != "npm.cmd" || len(args) != 2 || args[0] != "root" || args[1] != "-g" {
				return "", errors.New("unexpected runCmd call")
			}
			return root + "\r\n", nil
		},
	}
	got, reason := l.Find(config.Defaults())
	if got != hit {
		t.Fatalf("path = %q, want %q", got, hit)
	}
	if !strings.Contains(reason, "npm root") {
		t.Fatalf("reason = %q, want npm root mention", reason)
	}
}

func TestFindAllMissReturnsEmpty(t *testing.T) {
	l := &lsLookup{
		statFn: existsAt(),
		envFn:  envFromMap(map[string]string{}),
		runCmd: func(string, ...string) (string, error) { return "", errors.New("npm not installed") },
	}
	got, reason := l.Find(config.Defaults())
	if got != "" {
		t.Fatalf("path = %q, want empty", got)
	}
	if !strings.Contains(reason, "not found") {
		t.Fatalf("reason = %q, want 'not found'", reason)
	}
}

func TestBuildUsesNodeWhenLocated(t *testing.T) {
	hit := `C:\global\npm\@github\copilot-language-server\dist\language-server.js`
	l := &lsLookup{
		statFn: existsAt(hit),
		envFn:  envFromMap(map[string]string{}),
		runCmd: func(string, ...string) (string, error) { return `C:\global\npm` + "\n", nil },
	}
	cmd := l.build(config.Defaults(), &proxylog.Logger{})
	if cmd == nil {
		t.Fatal("build returned nil")
	}
	// argv[0] should resolve to a node binary; we accept either node.exe
	// (resolved) or the literal 'node.exe' (LookPath miss in this env).
	if !strings.Contains(strings.ToLower(cmd.Args[0]), "node") {
		t.Fatalf("argv[0] = %q, want 'node...'", cmd.Args[0])
	}
	if cmd.Args[1] != hit {
		t.Fatalf("argv[1] = %q, want js path %q", cmd.Args[1], hit)
	}
	if cmd.Args[2] != "--stdio" {
		t.Fatalf("argv[2] = %q, want --stdio", cmd.Args[2])
	}
}

func TestBuildFallsBackToNpxWhenNothingFound(t *testing.T) {
	// Make Find always miss AND make the install command fail fast.
	l := &lsLookup{
		statFn: existsAt(),
		envFn:  envFromMap(map[string]string{}),
		runCmd: func(string, ...string) (string, error) { return "", errors.New("no npm") },
	}
	// Override installGlobal indirectly by making npm.cmd unavailable —
	// we cannot patch the method directly, but the test still verifies
	// the structural fallback to npx by inspecting argv shape.
	// The build() flow will: Find -> miss, installGlobal -> error, Find again -> miss,
	// then fallback to npx.
	cmd := l.build(config.Defaults(), &proxylog.Logger{})
	if cmd == nil {
		t.Fatal("build returned nil")
	}
	if !strings.Contains(strings.ToLower(cmd.Args[0]), "npx") {
		t.Fatalf("argv[0] = %q, want 'npx...'", cmd.Args[0])
	}
	if cmd.Args[1] != "--yes" {
		t.Fatalf("argv[1] = %q, want --yes", cmd.Args[1])
	}
	if cmd.Args[2] != config.DefaultLanguageServerPackage {
		t.Fatalf("argv[2] = %q, want %q", cmd.Args[2], config.DefaultLanguageServerPackage)
	}
	if cmd.Args[3] != "--stdio" {
		t.Fatalf("argv[3] = %q, want --stdio", cmd.Args[3])
	}
}
