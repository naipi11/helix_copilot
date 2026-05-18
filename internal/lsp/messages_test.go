package lsp

import (
	"encoding/json"
	"testing"
)

func TestInitializeMessageAdvertisesInlineCompletion(t *testing.T) {
	msg := InitializeMessage("/work", "gpt-4o-copilot")
	var decoded map[string]any
	if err := json.Unmarshal([]byte(msg), &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded["method"] != "initialize" {
		t.Fatalf("method = %v", decoded["method"])
	}
	params := decoded["params"].(map[string]any)
	caps := params["capabilities"].(map[string]any)
	textDocument := caps["textDocument"].(map[string]any)
	inline := textDocument["inlineCompletion"].(map[string]any)
	if inline["dynamicRegistration"] != false {
		t.Fatalf("inlineCompletion.dynamicRegistration = %v", inline["dynamicRegistration"])
	}
	initOptions := params["initializationOptions"].(map[string]any)
	if initOptions["model"] != "gpt-4o-copilot" {
		t.Fatalf("model = %v", initOptions["model"])
	}
}

func TestFrameAddsContentLengthHeader(t *testing.T) {
	got := Frame(`{"jsonrpc":"2.0"}`)
	want := "Content-Length: 17\r\n\r\n{\"jsonrpc\":\"2.0\"}"
	if got != want {
		t.Fatalf("Frame() = %q, want %q", got, want)
	}
}
