package rpc

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRPCHandler(t *testing.T) {
	tests := []struct {
		name        string
		request     interface{}
		wantStatus  int
		wantMethod  string
		wantResult  interface{}
		wantError   *RPCError
		checkResult func(interface{}) bool
	}{
		{
			name: "valid custom method request",
			request: RPCRequest{
				JSONRPC: "2.0",
				Method:  "my_custom_method",
				Params:  json.RawMessage(`[]`),
				ID:      rawMessagePtr(json.RawMessage(`1`)),
			},
			wantStatus: http.StatusOK,
			wantMethod: "my_custom_method",
			checkResult: func(result interface{}) bool {
				m, ok := result.(map[string]interface{})
				if !ok {
					return false
				}
				return m["echo"] == "你好，Remix！这是 Go 模块的自定义返回。"
			},
		},
		{
			name: "eth_blockNumber method",
			request: RPCRequest{
				JSONRPC: "2.0",
				Method:  "eth_blockNumber",
				Params:  json.RawMessage(`[]`),
				ID:      rawMessagePtr(json.RawMessage(`2`)),
			},
			wantStatus: http.StatusOK,
			wantMethod: "eth_blockNumber",
			wantResult: "0x1a",
		},
		{
			name:       "invalid JSON",
			request:    "not a valid json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "method not found",
			request: RPCRequest{
				JSONRPC: "2.0",
				Method:  "unknown_method",
				Params:  json.RawMessage(`[]`),
				ID:      rawMessagePtr(json.RawMessage(`3`)),
			},
			wantStatus: http.StatusOK,
			wantError: &RPCError{
				Code:    -32601,
				Message: "Method not found",
			},
		},
		{
			name: "notification (no ID)",
			request: RPCRequest{
				JSONRPC: "2.0",
				Method:  "my_custom_method",
				Params:  json.RawMessage(`[]`),
				ID:      nil,
			},
			wantStatus: http.StatusOK,
			wantMethod: "my_custom_method",
			checkResult: func(result interface{}) bool {
				m, ok := result.(map[string]interface{})
				if !ok {
					return false
				}
				return m["echo"] == "你好，Remix！这是 Go 模块的自定义返回。"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			var err error

			switch v := tt.request.(type) {
			case string:
				reqBody = []byte(v)
			default:
				reqBody, err = json.Marshal(v)
				if err != nil {
					t.Fatalf("failed to marshal request: %v", err)
				}
			}

			req := httptest.NewRequest("POST", "/", bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			Handler(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, resp.StatusCode)
			}

			if tt.wantStatus != http.StatusOK {
				return
			}

			var rpcResp RPCResponse
			if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if tt.wantError != nil {
				if rpcResp.Error == nil {
					t.Error("expected error, got none")
				} else if rpcResp.Error.Code != tt.wantError.Code || rpcResp.Error.Message != tt.wantError.Message {
					t.Errorf("expected error %v, got %v", tt.wantError, rpcResp.Error)
				}
				return
			}

			if tt.checkResult != nil {
				if !tt.checkResult(rpcResp.Result) {
					t.Errorf("result check failed, got %v", rpcResp.Result)
				}
			} else if tt.wantResult != nil {
				if rpcResp.Result != tt.wantResult {
					t.Errorf("expected result %v, got %v", tt.wantResult, rpcResp.Result)
				}
			}

			// Verify JSON-RPC version
			if rpcResp.JSONRPC != "2.0" {
				t.Errorf("expected jsonrpc version 2.0, got %s", rpcResp.JSONRPC)
			}
		})
	}
}

func rawMessagePtr(msg json.RawMessage) *json.RawMessage {
	return &msg
}

func TestInvalidRequestBody(t *testing.T) {
	// 模拟读取请求体失败的情况
	req := httptest.NewRequest("POST", "/", &errorReader{})
	w := httptest.NewRecorder()

	Handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

// errorReader 是一个总是返回错误的 io.Reader 实现
type errorReader struct{}

func (er *errorReader) Read(p []byte) (n int, err error) {
	return 0, http.ErrBodyReadAfterClose
}
