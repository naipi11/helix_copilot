// Package proxylog provides a lightweight file-only logger for the LSP proxy.
//
// The proxy talks LSP over stdout, so logging to stdout would corrupt the
// protocol frames. Helix also swallows the proxy's stderr by default, so
// stderr is not a reliable diagnostic channel either. This package writes
// diagnostics to a single file the user can tail when debugging.
//
// Disable with HELIX_COPILOT_LOG=0 (or "off"). Default is enabled.
package proxylog

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Logger writes diagnostics to a single file. It is safe to use a nil-backed
// Logger: every method becomes a silent no-op.
type Logger struct {
	mu   sync.Mutex
	out  *log.Logger // may be nil if no file could be opened
	file *os.File    // may be nil if disabled or open failed
	path string
}

// New constructs a Logger by inspecting HELIX_COPILOT_LOG and opening the
// platform-specific log file. It never returns nil and never panics.
// If the log file cannot be opened, the returned Logger silently drops writes.
func New() *Logger {
	if disabled() {
		return &Logger{}
	}
	dir := defaultLogDir()
	if dir == "" {
		return &Logger{}
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return &Logger{}
	}
	return newAt(filepath.Join(dir, "proxy.log"))
}

// newAt opens the given log file path, truncating it. Internal seam exposed
// for tests so they can write into a temporary directory.
func newAt(path string) *Logger {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return &Logger{}
	}
	return &Logger{
		out:  log.New(f, "", log.Ldate|log.Ltime|log.Lmicroseconds),
		file: f,
		path: path,
	}
}

// Printf records a single line into the log file. No-op when disabled.
func (l *Logger) Printf(format string, v ...any) {
	if l == nil || l.out == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out.Output(2, fmt.Sprintf(format, v...))
}

// Path returns the absolute path of the log file, or "" when disabled.
func (l *Logger) Path() string {
	if l == nil {
		return ""
	}
	return l.path
}

// Close releases the underlying file handle. Safe on a nil-backed Logger.
func (l *Logger) Close() error {
	if l == nil || l.file == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	err := l.file.Close()
	l.out = nil
	l.file = nil
	return err
}

// Tee returns an io.Writer that prefixes each written line with the given
// label and forwards it into the log file. Used for child stderr passthrough.
// Returns io.Discard when the logger is disabled.
func (l *Logger) Tee(prefix string) io.Writer {
	if l == nil || l.out == nil {
		return io.Discard
	}
	return &teeWriter{logger: l, prefix: prefix}
}

type teeWriter struct {
	logger *Logger
	prefix string
	buf    bytes.Buffer
	mu     sync.Mutex
}

func (t *teeWriter) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	n, _ := t.buf.Write(p)
	for {
		idx := bytes.IndexByte(t.buf.Bytes(), '\n')
		if idx < 0 {
			break
		}
		line := strings.TrimRight(string(t.buf.Next(idx+1)), "\r\n")
		if line != "" {
			t.logger.Printf("%s %s", t.prefix, line)
		}
	}
	return n, nil
}

func disabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("HELIX_COPILOT_LOG")))
	return v == "0" || v == "off" || v == "false" || v == "no"
}

func defaultLogDir() string {
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "helix-copilot")
		}
		if profile := os.Getenv("USERPROFILE"); profile != "" {
			return filepath.Join(profile, "helix-copilot")
		}
		return ""
	}
	if state := os.Getenv("XDG_STATE_HOME"); state != "" {
		return filepath.Join(state, "helix-copilot")
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".local", "state", "helix-copilot")
	}
	return ""
}
