package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/naipi11/helix_copilot/internal/config"
)

// Proxy bridges Helix LSP (stdin/stdout) to Copilot language server.
// It translates textDocument/completion ↔ textDocument/inlineCompletion.
type Proxy struct {
	childCmd    *exec.Cmd
	childIn     io.WriteCloser
	childOut    io.ReadCloser
	childErr    io.ReadCloser
	helixReader *bufio.Reader

	mu        sync.Mutex
	idMapping map[int]json.RawMessage // inlineCompletion ID → original completion ID
	nextID    int
	initReqID json.RawMessage
	ready     bool // Copilot LS has finished initializing
}

func NewProxy() *Proxy {
	return &Proxy{
		idMapping: make(map[int]json.RawMessage),
		nextID:    10000,
	}
}

func (p *Proxy) Run() error {
	cmdArgs := []string{"--yes", config.DefaultLanguageServerPackage, "--stdio"}
	cmd := exec.Command("npx", cmdArgs...)

	childIn, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("child stdin pipe: %w", err)
	}
	childOut, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("child stdout pipe: %w", err)
	}
	childErr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("child stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start copilot LS: %w", err)
	}

	p.childCmd = cmd
	p.childIn = childIn
	p.childOut = childOut
	p.childErr = childErr
	p.helixReader = bufio.NewReader(os.Stdin)

	// Forward child stderr to our stderr
	go func() {
		_, _ = io.Copy(os.Stderr, childErr)
	}()

	// Signal channel for goroutine lifecycle
	done := make(chan struct{})
	defer close(done)

	// Forward child stdout → our stdout (the return direction)
	go p.forwardChildToHelix(done)

	// Forward our stdin (Helix) → child stdin (main goroutine)
	return p.forwardHelixToChild()
}

func (p *Proxy) forwardHelixToChild() error {
	for {
		msg, err := readLSP(p.helixReader)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read from helix: %w", err)
		}

		method := rawMethod(msg)

		switch {
		case method == "initialize":
			if err := p.handleInitialize(msg); err != nil {
				return err
			}
		case method == "textDocument/completion":
			if !p.ready {
				// Copilot LS not ready yet; drop silently.
				// Helix will retry on next keystroke.
				continue
			}
			if err := p.handleCompletionRequest(msg); err != nil {
				fmt.Fprintf(os.Stderr, "completion bridge error: %v\n", err)
				// Fallback: forward as-is
				if err := writeLSP(p.childIn, msg); err != nil {
					return err
				}
			}
		default:
			if err := writeLSP(p.childIn, msg); err != nil {
				return err
			}
		}
	}
}

func (p *Proxy) forwardChildToHelix(done chan struct{}) {
	reader := bufio.NewReader(p.childOut)
	for {
		msg, err := readLSP(reader)
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Fprintf(os.Stderr, "read from child: %v\n", err)
			return
		}

		id, hasID := rawNumericID(msg)
		if hasID {
			p.mu.Lock()
			origID, mapped := p.idMapping[id]
			if mapped {
				delete(p.idMapping, id)
			}
			p.mu.Unlock()

			if mapped {
				// This is an inlineCompletion response; rewrite as completion response
				modified := p.rewriteInlineResponseAsCompletion(msg, origID)
				if err := writeLSP(os.Stdout, modified); err != nil {
					fmt.Fprintf(os.Stderr, "write to helix: %v\n", err)
					return
				}
				continue
			}

			// Check if this is an initialize response → inject completionProvider
			initResp, isInit := p.tryParseInitResponse(msg)
			if isInit {
				p.ready = true
				modified := injectCompletionProvider(initResp)
				if err := writeLSP(os.Stdout, modified); err != nil {
					fmt.Fprintf(os.Stderr, "write to helix: %v\n", err)
					return
				}
				continue
			}
		}

		// Forward unchanged
		if err := writeLSP(os.Stdout, msg); err != nil {
			fmt.Fprintf(os.Stderr, "write to helix: %v\n", err)
			return
		}
	}
}

func (p *Proxy) handleInitialize(msg []byte) error {
	// Inject editorInfo and editorPluginInfo into initializationOptions
	// (Copilot LS requires these)
	var req map[string]any
	if err := json.Unmarshal(msg, &req); err == nil {
		params, _ := req["params"].(map[string]any)
		if params != nil {
			initOpts, _ := params["initializationOptions"].(map[string]any)
			if initOpts == nil {
				initOpts = make(map[string]any)
				params["initializationOptions"] = initOpts
			}
			if _, exists := initOpts["editorInfo"]; !exists {
				initOpts["editorInfo"] = map[string]any{
					"name":    "helix",
					"version": "25.07.1",
				}
			}
			if _, exists := initOpts["editorPluginInfo"]; !exists {
				initOpts["editorPluginInfo"] = map[string]any{
					"name":    "helix-copilot",
					"version": "0.1.0",
				}
			}
			modified, _ := json.Marshal(req)
			return writeLSP(p.childIn, modified)
		}
	}
	// Fallback: forward as-is
	return writeLSP(p.childIn, msg)
}

func (p *Proxy) handleCompletionRequest(msg []byte) error {
	// Parse the completion request
	var req map[string]any
	if err := json.Unmarshal(msg, &req); err != nil {
		return err
	}

	// Extract the ID
	origID, ok := rawID(msg)
	if !ok {
		return fmt.Errorf("completion request missing id")
	}

	// Extract params
	params, ok := req["params"].(map[string]any)
	if !ok {
		return fmt.Errorf("completion request missing params")
	}

	// Create inline completion params from completion params
	inlineParams := map[string]any{
		"textDocument": params["textDocument"],
		"position":     params["position"],
		"context": map[string]any{
			"triggerKind": 1, // Invoked
		},
	}

	// Generate a new ID for the inline request
	p.mu.Lock()
	newID := p.nextID
	p.nextID++
	p.idMapping[newID] = origID
	p.mu.Unlock()

	inlineReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      newID,
		"method":  "textDocument/inlineCompletion",
		"params":  inlineParams,
	}

	data, err := json.Marshal(inlineReq)
	if err != nil {
		return err
	}

	return writeLSP(p.childIn, data)
}

func (p *Proxy) rewriteInlineResponseAsCompletion(msg []byte, origID json.RawMessage) []byte {
	var resp map[string]any
	if err := json.Unmarshal(msg, &resp); err != nil {
		return msg // fallback: return as-is
	}

	// Build a completion response
	completionResp := map[string]any{
		"jsonrpc": "2.0",
	}
	completionResp["id"] = json.RawMessage(origID)

	// If it's an error, forward the error
	if err, ok := resp["error"]; ok {
		completionResp["error"] = err
		data, _ := json.Marshal(completionResp)
		return data
	}

	// Extract inline completion items. LSP allows either an array of
	// InlineCompletionItem or an InlineCompletionList { items = [...] }.
	var items []any
	switch result := resp["result"].(type) {
	case []any:
		items = result
	case map[string]any:
		items, _ = result["items"].([]any)
	}
	if len(items) == 0 {
		completionResp["result"] = nil
		data, _ := json.Marshal(completionResp)
		return data
	}

	// Convert InlineCompletionItems to CompletionItems
	var completionItems []map[string]any
	for _, item := range items {
		inlineItem, ok := item.(map[string]any)
		if !ok {
			continue
		}

		insertText, _ := inlineItem["insertText"].(string)
		if insertText == "" {
			insertText, _ = inlineItem["text"].(string)
		}
		if insertText == "" {
			continue
		}

		completionItems = append(completionItems, map[string]any{
			"label":            truncateFirstLine(insertText),
			"insertText":       insertText,
			"detail":           "GitHub Copilot",
			"insertTextFormat": 1, // PlainText
		})
	}

	if len(completionItems) == 0 {
		completionResp["result"] = nil
	} else {
		completionResp["result"] = map[string]any{
			"isIncomplete": false,
			"items":        completionItems,
		}
	}

	data, _ := json.Marshal(completionResp)
	return data
}

func truncateFirstLine(s string) string {
	idx := strings.IndexAny(s, "\n\r")
	if idx >= 0 {
		if idx > 80 {
			return s[:idx]
		}
		return s[:idx]
	}
	if len(s) > 80 {
		return s[:80]
	}
	return s
}

// injectCompletionProvider adds a basic completionProvider capability to
// the Copilot LS initialize response so Helix sends completion requests.
func injectCompletionProvider(msg []byte) []byte {
	var resp map[string]any
	if err := json.Unmarshal(msg, &resp); err != nil {
		return msg
	}

	result, ok := resp["result"].(map[string]any)
	if !ok {
		return msg
	}

	caps, ok := result["capabilities"].(map[string]any)
	if !ok {
		return msg
	}

	// Inject standard completion provider
	if _, exists := caps["completionProvider"]; !exists {
		caps["completionProvider"] = map[string]any{
			"triggerCharacters": []string{".", " "},
			"resolveProvider":   false,
		}
	}
	// Inject inline completion provider (for native ghost text support)
	if _, exists := caps["inlineCompletionProvider"]; !exists {
		caps["inlineCompletionProvider"] = map[string]any{}
	}

	data, _ := json.Marshal(resp)
	return data
}

// ----- JSON-RPC LSP protocol helpers -----

func readLSP(r *bufio.Reader) ([]byte, error) {
	contentLength := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[0]), "Content-Length") {
			contentLength, err = strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Length: %w", err)
			}
		}
	}
	if contentLength < 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}
	body := make([]byte, contentLength)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	return body, nil
}

func writeLSP(w io.Writer, data []byte) error {
	var frame strings.Builder
	fmt.Fprintf(&frame, "Content-Length: %d\r\n\r\n", len(data))
	frame.Write(data)
	_, err := io.WriteString(w, frame.String())
	return err
}

func rawMethod(msg []byte) string {
	var parsed struct {
		Method string `json:"method"`
	}
	if err := json.Unmarshal(msg, &parsed); err != nil {
		return ""
	}
	return parsed.Method
}

func rawID(msg []byte) (json.RawMessage, bool) {
	var parsed struct {
		ID json.RawMessage `json:"id"`
	}
	if err := json.Unmarshal(msg, &parsed); err != nil || len(parsed.ID) == 0 || string(parsed.ID) == "null" {
		return nil, false
	}
	return append(json.RawMessage(nil), parsed.ID...), true
}

func rawNumericID(msg []byte) (int, bool) {
	id, ok := rawID(msg)
	if !ok {
		return 0, false
	}
	var idInt int
	if err := json.Unmarshal(id, &idInt); err == nil {
		return idInt, true
	}
	var idFloat float64
	if err := json.Unmarshal(id, &idFloat); err == nil {
		return int(idFloat), true
	}
	return 0, false
}

func (p *Proxy) tryParseInitResponse(msg []byte) ([]byte, bool) {
	id, hasID := rawID(msg)
	if !hasID {
		return nil, false
	}
	if len(p.initReqID) == 0 {
		p.initReqID = json.RawMessage(`1`)
	}
	if string(id) != string(p.initReqID) {
		return nil, false
	}
	var resp struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(msg, &resp); err != nil {
		return nil, false
	}
	// Check if this looks like an initialize result with capabilities
	var result struct {
		Capabilities map[string]any `json:"capabilities"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, false
	}
	if result.Capabilities == nil {
		return nil, false
	}
	return msg, true
}

func tryParseInitResponse(msg []byte) ([]byte, bool) {
	return (&Proxy{initReqID: json.RawMessage(`1`)}).tryParseInitResponse(msg)
}
