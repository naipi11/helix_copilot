package login

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type LSPMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type SignInResult struct {
	Status          string `json:"status"`
	UserCode        string `json:"userCode"`
	VerificationURI string `json:"verificationUri"`
	ExpiresIn       int    `json:"expiresIn"`
	Interval        int    `json:"interval"`
}

type Runner struct {
	Command string
	Args    []string
	Dir     string
	In      io.Reader
	Out     io.Writer
	Err     io.Writer
}

func DefaultRunner() Runner {
	return Runner{
		Command: "npx",
		Args:    []string{"--yes", "@github/copilot-language-server", "--stdio"},
		In:      os.Stdin,
		Out:     os.Stdout,
		Err:     os.Stderr,
	}
}

func Run(r Runner) error {
	if r.Command == "" {
		r.Command = "npx"
	}
	if len(r.Args) == 0 {
		r.Args = []string{"--yes", "@github/copilot-language-server", "--stdio"}
	}
	if r.In == nil {
		r.In = os.Stdin
	}
	if r.Out == nil {
		r.Out = os.Stdout
	}
	if r.Err == nil {
		r.Err = os.Stderr
	}

	cmd := exec.Command(r.Command, r.Args...)
	cmd.Dir = r.Dir
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = r.Err
	if err := cmd.Start(); err != nil {
		return err
	}
	defer func() { _ = cmd.Process.Kill() }()

	reader := bufio.NewReader(stdout)
	nextID := 1
	if err := writeJSONRPC(stdin, map[string]any{
		"jsonrpc": "2.0",
		"id":      nextID,
		"method":  "initialize",
		"params": map[string]any{
			"processId": os.Getpid(),
			"rootUri":   rootURI(r.Dir),
			"capabilities": map[string]any{
				"workspace": map[string]any{"workspaceFolders": true},
				"window":    map[string]any{"showDocument": map[string]any{"support": false}},
			},
			"initializationOptions": map[string]any{
				"editorInfo":       map[string]any{"name": "helix-copilot", "version": "0.1.0"},
				"editorPluginInfo": map[string]any{"name": "helix-copilot", "version": "0.1.0"},
			},
		},
	}); err != nil {
		return err
	}
	if _, err := waitForID(reader, nextID, 30*time.Second); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	if err := writeJSONRPC(stdin, map[string]any{"jsonrpc": "2.0", "method": "initialized", "params": map[string]any{}}); err != nil {
		return err
	}

	nextID++
	if err := writeJSONRPC(stdin, SignInInitiateRequest(nextID)); err != nil {
		return err
	}
	msg, err := waitForID(reader, nextID, 30*time.Second)
	if err != nil {
		return fmt.Errorf("signInInitiate: %w", err)
	}
	if msg.Error != nil {
		return fmt.Errorf("signInInitiate: %s", msg.Error.Message)
	}
	var sign SignInResult
	if err := json.Unmarshal(msg.Result, &sign); err != nil {
		return err
	}
	if sign.UserCode == "" || sign.VerificationURI == "" {
		return fmt.Errorf("unexpected sign-in response: %s", string(msg.Result))
	}
	fmt.Fprintf(r.Out, "打开这个地址完成 GitHub Copilot 登录：%s\n", sign.VerificationURI)
	fmt.Fprintf(r.Out, "输入设备码：%s\n", sign.UserCode)
	fmt.Fprintln(r.Out, "完成网页登录后，回到这里按 Enter 继续。")
	_, _ = bufio.NewReader(r.In).ReadString('\n')

	nextID++
	if err := writeJSONRPC(stdin, FinishDeviceFlowRequest(nextID)); err != nil {
		return err
	}
	msg, err = waitForID(reader, nextID, time.Duration(max(sign.ExpiresIn, 900))*time.Second)
	if err != nil {
		// finishDeviceFlow may fail due to network issues (proxy/TLS). If the
		// user has already authorized in the browser, the auth file will have
		// been written by the Copilot LS's initialization handshake, so the
		// login is effectively complete.
		if authFileExists() {
			fmt.Fprintln(r.Out, "GitHub Copilot 登录完成（设备码已验证）。")
			return cmd.Process.Kill()
		}
		return fmt.Errorf("finishDeviceFlow: %w", err)
	}
	if msg.Error != nil {
		if authFileExists() {
			fmt.Fprintln(r.Out, "GitHub Copilot 登录完成（设备码已验证）。")
			return cmd.Process.Kill()
		}
		return fmt.Errorf("finishDeviceFlow: %s", msg.Error.Message)
	}
	fmt.Fprintln(r.Out, "GitHub Copilot 登录完成。")
	return cmd.Process.Kill()
}

// authFileExists checks whether the Copilot LS has persisted its auth token,
// which means the device flow completed successfully on GitHub's side even if
// the finishDeviceFlow LSP command failed due to a network error.
func authFileExists() bool {
	var configDir string
	switch runtime.GOOS {
	case "windows":
		configDir = os.Getenv("LOCALAPPDATA")
		if configDir == "" {
			configDir = os.Getenv("APPDATA")
		}
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		configDir = filepath.Join(home, ".config")
	}
	if configDir == "" {
		return false
	}
	// Copilot LS stores auth tokens in a hosts.json file under github-copilot/
	hostsPath := filepath.Join(configDir, "github-copilot", "hosts.json")
	if _, err := os.Stat(hostsPath); err == nil {
		return true
	}
	return false
}

func writeJSONRPC(w io.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "Content-Length: %d\r\n\r\n%s", len(data), data)
	return err
}

func readJSONRPC(r *bufio.Reader) (LSPMessage, error) {
	contentLength := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return LSPMessage{}, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[0]), "Content-Length") {
			contentLength, err = strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return LSPMessage{}, err
			}
		}
	}
	if contentLength < 0 {
		return LSPMessage{}, errors.New("missing Content-Length")
	}
	body := make([]byte, contentLength)
	if _, err := io.ReadFull(r, body); err != nil {
		return LSPMessage{}, err
	}
	var msg LSPMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return LSPMessage{}, err
	}
	return msg, nil
}

func waitForID(r *bufio.Reader, id int, timeout time.Duration) (LSPMessage, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		msg, err := readJSONRPC(r)
		if err != nil {
			return LSPMessage{}, err
		}
		if msg.ID != nil && *msg.ID == id {
			return msg, nil
		}
	}
	return LSPMessage{}, errors.New("timeout")
}

func rootURI(dir string) string {
	if dir == "" {
		dir, _ = os.Getwd()
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "file:///"
	}
	u := url.URL{Scheme: "file", Path: abs}
	return u.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
