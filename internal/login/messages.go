package login

type Request struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

type ExecuteCommandRequest struct {
	JSONRPC string               `json:"jsonrpc"`
	ID      int                  `json:"id"`
	Method  string               `json:"method"`
	Params  ExecuteCommandParams `json:"params"`
}

type ExecuteCommandParams struct {
	Command   string `json:"command"`
	Arguments []any  `json:"arguments"`
}

func SignInInitiateRequest(id int) Request {
	return Request{JSONRPC: "2.0", ID: id, Method: "signInInitiate", Params: map[string]any{}}
}

func FinishDeviceFlowRequest(id int) ExecuteCommandRequest {
	return ExecuteCommandRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  "workspace/executeCommand",
		Params: ExecuteCommandParams{
			Command:   "github.copilot.finishDeviceFlow",
			Arguments: []any{},
		},
	}
}
