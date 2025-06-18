package api

import (
	"ablockchain/core"
	"encoding/json"
	"net/http"
)

type JSONRPCRequest struct {
	JSONRPC string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
	ID      int               `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      int         `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type API struct {
	bc *core.Blockchain
}

func NewAPI(bc *core.Blockchain) *API {
	return &API{bc: bc}
}

func (a *API) Handler(w http.ResponseWriter, r *http.Request) {
	// 添加 CORS 头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	// 对于浏览器的预检请求，直接返回 200
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, req.ID, -32700, "Invalid JSON")
		return
	}

	var result interface{}
	var errResp *RPCError

	switch req.Method {
	case "getChain":
		var err error
		result, err = a.bc.GetChain()
		if err != nil {
			errResp = &RPCError{-32000, "Failed to get chain"}
		}
	case "getBlockByIndex":
		if len(req.Params) != 1 {
			errResp = &RPCError{-32602, "Expected 1 parameter"}
			break
		}
		var index int
		if err := json.Unmarshal(req.Params[0], &index); err != nil {
			errResp = &RPCError{-32602, "Invalid parameter"}
			break
		}
		block, err := a.bc.GetBlockByIndex(index)
		if err != nil {
			errResp = &RPCError{-32000, "Block not found"}
			break
		}
		result = block

	default:
		errResp = &RPCError{-32601, "Method not found"}
	}

	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	if errResp != nil {
		resp.Error = errResp
	} else {
		resp.Result = result
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, id int, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &RPCError{Code: code, Message: message},
	})
}
