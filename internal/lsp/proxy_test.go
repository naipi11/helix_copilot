package lsp

import (
	"bytes"
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
	if !ok || string(id) != "42" {
		t.Fatalf("rawID = %s, %v", id, ok)
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

func TestRewriteInlineResponseAsCompletionHandlesArrayResult(t *testing.T) {
	proxy := NewProxy()
	msg := []byte(`{"jsonrpc":"2.0","id":10000,"result":[{"insertText":"fmt.Println(\"hi\")"},{"insertText":"return nil"}]}`)
	modified := proxy.rewriteInlineResponseAsCompletion(msg, json.RawMessage(`7`))

	var resp map[string]any
	if err := json.Unmarshal(modified, &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp["id"].(float64) != 7 {
		t.Fatalf("id = %v, want 7", resp["id"])
	}
	result := resp["result"].(map[string]any)
	items := result["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("items len = %d, want 2", len(items))
	}
	first := items[0].(map[string]any)
	if first["insertText"] != "fmt.Println(\"hi\")" {
		t.Fatalf("insertText = %v", first["insertText"])
	}
}

func TestRewriteInlineResponseAsCompletionUsesTextFallback(t *testing.T) {
	proxy := NewProxy()
	msg := []byte(`{"jsonrpc":"2.0","id":10000,"result":{"items":[{"text":"legacy text completion"}]}}`)
	modified := proxy.rewriteInlineResponseAsCompletion(msg, json.RawMessage(`7`))

	var resp map[string]any
	if err := json.Unmarshal(modified, &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	result := resp["result"].(map[string]any)
	items := result["items"].([]any)
	first := items[0].(map[string]any)
	if first["insertText"] != "legacy text completion" {
		t.Fatalf("insertText = %v", first["insertText"])
	}
}

func TestHandleCompletionRequestPreservesStringOriginalID(t *testing.T) {
	proxy := NewProxy()
	var out testWriteCloser
	proxy.childIn = &out
	msg := []byte(`{"jsonrpc":"2.0","id":"req-42","method":"textDocument/completion","params":{"textDocument":{"uri":"file:///tmp/a.go"},"position":{"line":0,"character":1}}}`)
	if err := proxy.handleCompletionRequest(msg); err != nil {
		t.Fatalf("handleCompletionRequest: %v", err)
	}
	origID, ok := proxy.idMapping[10000]
	if !ok {
		t.Fatal("missing id mapping for generated inline request")
	}
	if string(origID) != `"req-42"` {
		t.Fatalf("mapped original id = %s", origID)
	}
	modified := proxy.rewriteInlineResponseAsCompletion([]byte(`{"jsonrpc":"2.0","id":10000,"result":{"items":[{"insertText":"ok"}]}}`), origID)
	var resp map[string]any
	if err := json.Unmarshal(modified, &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp["id"] != "req-42" {
		t.Fatalf("response id = %v", resp["id"])
	}
}

func TestProxyRecognizesInitializeResponseWithNonOneID(t *testing.T) {
	proxy := NewProxy()
	proxy.initReqID = json.RawMessage(`5`)
	msg := []byte(`{"jsonrpc":"2.0","id":5,"result":{"capabilities":{"inlineCompletionProvider":{}}}}`)
	_, ok := proxy.tryParseInitResponse(msg)
	if !ok {
		t.Fatal("should recognize initialize response with recorded non-1 id")
	}
}

func TestWriteLSPWritesSingleFrame(t *testing.T) {
	var out testWriteCloser
	if err := writeLSP(&out, []byte(`{"ok":true}`)); err != nil {
		t.Fatalf("writeLSP: %v", err)
	}
	want := "Content-Length: 11\r\n\r\n{\"ok\":true}"
	if out.String() != want {
		t.Fatalf("frame = %q, want %q", out.String(), want)
	}
}

type testWriteCloser struct{ bytes.Buffer }

func (t *testWriteCloser) Close() error { return nil }
