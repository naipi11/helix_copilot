package lsp

import (
	"encoding/json"
	"fmt"
)

type request struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

func Frame(payload string) string {
	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len([]byte(payload)), payload)
}

func InitializeMessage(rootURI, model string) string {
	msg := request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]any{
			"processId": nil,
			"rootUri":   rootURI,
			"capabilities": map[string]any{
				"textDocument": map[string]any{
					"inlineCompletion": map[string]any{
						"dynamicRegistration": false,
					},
					"completion": map[string]any{
						"completionItem": map[string]any{
							"snippetSupport": true,
						},
					},
				},
				"workspace": map[string]any{
					"workspaceFolders": true,
				},
			},
			"initializationOptions": map[string]any{
				"model": model,
			},
		},
	}
	data, _ := json.Marshal(msg)
	return string(data)
}
