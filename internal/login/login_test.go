package login

import "testing"

func TestSignInInitiateRequestUsesCopilotMethod(t *testing.T) {
	got := SignInInitiateRequest(42)
	if got.JSONRPC != "2.0" {
		t.Fatalf("JSONRPC = %q", got.JSONRPC)
	}
	if got.ID != 42 {
		t.Fatalf("ID = %d", got.ID)
	}
	if got.Method != "signInInitiate" {
		t.Fatalf("Method = %q, want signInInitiate", got.Method)
	}
	if got.Params == nil {
		t.Fatalf("Params should be an empty object, got nil")
	}
}

func TestFinishDeviceFlowRequestUsesExecuteCommand(t *testing.T) {
	got := FinishDeviceFlowRequest(99)
	if got.Method != "workspace/executeCommand" {
		t.Fatalf("Method = %q", got.Method)
	}
	if got.Params.Command != "github.copilot.finishDeviceFlow" {
		t.Fatalf("Command = %q", got.Params.Command)
	}
	if len(got.Params.Arguments) != 0 {
		t.Fatalf("Arguments = %#v, want empty", got.Params.Arguments)
	}
}
