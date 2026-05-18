package lsp

import (
	"encoding/json"
	"testing"
)

func TestInjectCompletionProvider(t *testing.T) {
	original := []byte(`{"jsonrpc":"2.0","id":1,"result":{"capabilities":{"inlineCompletionProvider":{},"textDocumentSync":{"openClose":true},"workspace":{"workspaceFolders":true}}}}`)
	modified := injectCompletionProvider(original)

	var resp map[string]any
	if err := json.Unmarshal(modified, &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	result := resp["result"].(map[string]any)
	caps := result["capabilities"].(map[string]any)

	if _, ok := caps["completionProvider"]; !ok {
		t.Fatal("injected response missing completionProvider")
	}
	if _, ok := caps["inlineCompletionProvider"]; !ok {
		t.Fatal("injected response lost inlineCompletionProvider")
	}
}

func TestInjectCompletionProviderSkipsIfAlreadyPresent(t *testing.T) {
	original := []byte(`{"jsonrpc":"2.0","id":1,"result":{"capabilities":{"completionProvider":{"triggerCharacters":["."]},"inlineCompletionProvider":{}}}}`)
	modified := injectCompletionProvider(original)

	var resp map[string]any
	json.Unmarshal(modified, &resp)
	caps := resp["result"].(map[string]any)["capabilities"].(map[string]any)
	cp := caps["completionProvider"]
	if cp == nil {
		t.Fatal("completionProvider is nil after re-marshal")
	}
	cpMap, ok := cp.(map[string]any)
	if !ok {
		t.Fatalf("completionProvider is %T, not map", cp)
	}
	chars, ok := cpMap["triggerCharacters"]
	if !ok {
		t.Fatal("triggerCharacters missing")
	}
	charsList := chars.([]any)
	if len(charsList) != 1 || charsList[0] != "." {
		t.Fatalf("unexpected triggerCharacters: %v", charsList)
	}
}

func TestTruncateFirstLine(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"fmt.Println", "fmt.Println"},
		{"package main\n\nfunc main()", "package main"},
	}
	for _, tc := range tests {
		got := truncateFirstLine(tc.input)
		if got != tc.want {
			t.Fatalf("truncateFirstLine(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
	// Long line should be truncated to <=80 chars
	long := "a very long line that exceeds eighty characters so it needs to be truncated nicely by the truncation helper function"
	got := truncateFirstLine(long)
	if len(got) > 80 {
		t.Fatalf("truncated line length = %d, want <=80", len(got))
	}
	if len(got) > len(long) {
		t.Fatalf("truncated should not be longer than original")
	}
}

func TestRawMethod(t *testing.T) {
	msg := []byte(`{"jsonrpc":"2.0","method":"textDocument/completion","id":3}`)
	if got := rawMethod(msg); got != "textDocument/completion" {
		t.Fatalf("rawMethod = %q", got)
	}
	msg2 := []byte(`{"jsonrpc":"2.0","id":1,"result":null}`)
	if got := rawMethod(msg2); got != "" {
		t.Fatalf("rawMethod should be empty for responses: %q", got)
	}
}

func TestRawID(t *testing.T) {
	msg := []byte(`{"jsonrpc":"2.0","id":42,"result":null}`)
	id, ok := rawID(msg)
	if !ok || id != 42 {
		t.Fatalf("rawID = %d, %v", id, ok)
	}
	msg2 := []byte(`{"jsonrpc":"2.0","method":"initialized","params":{}}`)
	_, ok = rawID(msg2)
	if ok {
		t.Fatal("notification should have no ID")
	}
}

func TestTryParseInitResponse(t *testing.T) {
	msg := []byte(`{"jsonrpc":"2.0","id":1,"result":{"capabilities":{"textDocumentSync":{"openClose":true}}}}`)
	_, ok := tryParseInitResponse(msg)
	if !ok {
		t.Fatal("should recognize init response")
	}

	msg2 := []byte(`{"jsonrpc":"2.0","id":2,"result":{"capabilities":{}}}`)
	_, ok = tryParseInitResponse(msg2)
	if ok {
		t.Fatal("id=2 should not be init response")
	}
}
