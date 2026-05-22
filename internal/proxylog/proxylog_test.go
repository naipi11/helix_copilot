package proxylog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewWritesToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "proxy.log")
	l := newAt(path)
	defer l.Close()

	if l.Path() != path {
		t.Fatalf("Path() = %q, want %q", l.Path(), path)
	}

	l.Printf("hello %s", "world")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if !strings.Contains(string(data), "hello world") {
		t.Fatalf("log file does not contain expected message:\n%s", string(data))
	}
}

func TestNewDisabledIsNoop(t *testing.T) {
	t.Setenv("HELIX_COPILOT_LOG", "0")
	l := New()
	defer l.Close()
	if l.Path() != "" {
		t.Fatalf("Path() = %q, want empty when disabled", l.Path())
	}
	// Should not panic and produce no output anywhere.
	l.Printf("should not appear")
}

func TestNewAtUnwritablePathIsNoop(t *testing.T) {
	dir := t.TempDir()
	// A path whose parent directory is actually a file, not a directory —
	// guaranteed-unwritable across platforms.
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("seed blocker: %v", err)
	}
	bad := filepath.Join(blocker, "proxy.log")
	l := newAt(bad)
	defer l.Close()
	if l.Path() != "" {
		t.Fatalf("Path() = %q, want empty when open failed", l.Path())
	}
	l.Printf("nothing should crash")
}

func TestTeePrefixesAndForwards(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "proxy.log")
	l := newAt(path)
	defer l.Close()

	tee := l.Tee("[child]")
	if _, err := tee.Write([]byte("line1\nline2\n")); err != nil {
		t.Fatalf("tee write: %v", err)
	}
	// Partial line should buffer until newline arrives.
	if _, err := tee.Write([]byte("partial")); err != nil {
		t.Fatalf("tee write: %v", err)
	}
	if _, err := tee.Write([]byte("-tail\n")); err != nil {
		t.Fatalf("tee write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	got := string(data)
	for _, want := range []string{"[child] line1", "[child] line2", "[child] partial-tail"} {
		if !strings.Contains(got, want) {
			t.Fatalf("log missing %q:\n%s", want, got)
		}
	}
}

func TestNilLoggerIsSafe(t *testing.T) {
	var l *Logger
	l.Printf("safe")
	if l.Path() != "" {
		t.Fatalf("nil logger Path should be empty")
	}
	if err := l.Close(); err != nil {
		t.Fatalf("nil logger Close: %v", err)
	}
	tee := l.Tee("[x]")
	if _, err := tee.Write([]byte("ignored\n")); err != nil {
		t.Fatalf("nil logger tee write: %v", err)
	}
}
