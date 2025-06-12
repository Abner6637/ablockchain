package system

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type RPCRequest struct {
	JSONRPC string           `json:"jsonrpc"`
	Method  string           `json:"method"`
	Params  json.RawMessage  `json:"params"`
	ID      *json.RawMessage `json:"id"`
}

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type RPCResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	Result  interface{}      `json:"result,omitempty"`
	Error   *RPCError        `json:"error,omitempty"`
	ID      *json.RawMessage `json:"id"`
}

func (sys *System) Handler(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Invalid or empty body", http.StatusBadRequest)
		return
	}

	var requests []RPCRequest
	isBatch := false

	if body[0] == '[' {
		if err := json.Unmarshal(body, &requests); err != nil {
			log.Printf("无法解析 JSON-RPC 批量请求: %s\n原始请求: %s", err, string(body))
			http.Error(w, "Invalid JSON-RPC batch", http.StatusBadRequest)
			return
		}
		isBatch = true
	} else {
		var req RPCRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Printf("无法解析 JSON-RPC 请求: %s\n原始请求: %s", err, string(body))
			http.Error(w, "Invalid JSON-RPC", http.StatusBadRequest)
			return
		}
		requests = append(requests, req)
	}

	var responses []RPCResponse
	for _, req := range requests {
		res := sys.handleSingleRequest(req)
		if res.ID != nil { // per spec: notifications (null id) should not be responded to
			responses = append(responses, res)
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if isBatch {
		if err := json.NewEncoder(w).Encode(responses); err != nil {
			log.Printf("响应编码失败: %s", err)
		}
	} else if len(responses) == 1 {
		if err := json.NewEncoder(w).Encode(responses[0]); err != nil {
			log.Printf("响应编码失败: %s", err)
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Unexpected server condition"))
	}
}

func (sys *System) handleSingleRequest(req RPCRequest) RPCResponse {
	fmt.Printf("**************************\n")
	fmt.Printf("Received method: %s\n", req.Method)
	fmt.Printf("Params: %s\n", string(req.Params))
	if req.ID != nil {
		fmt.Printf("ID: %s\n", string(*req.ID))
	}
	fmt.Printf("**************************\n")

	res := RPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	/*************** Accounts ***************/
	case "eth_requestAccounts":
		res.Result = sys.eth_requestAccounts()
	case "eth_accounts":
		res.Result = sys.eth_accounts()
	case "eth_getBalance":
		var params []interface{}
		if err := json.Unmarshal(req.Params, &params); err != nil || len(params) < 1 {
			res.Error = &RPCError{Code: -32602, Message: "Invalid params"}
		} else {
			acc := params[0].(string)
			res.Result = sys.eth_getBalance(acc)
		}
	case "eth_sign":
		res.Result = "1"
	case "eth_chainId":
		res.Result = sys.eth_chainId()
	case "eth_signTypedData":
		res.Result = "1"
	case "eth_signTypedData_v4":
		res.Result = "1"
	case "eth_getPKey":
		res.Result = "1"

	/*************** Blocks ***************/
	case "eth_getBlockByNumber":
		var params []interface{}
		if err := json.Unmarshal(req.Params, &params); err != nil || len(params) < 1 {
			res.Error = &RPCError{Code: -32602, Message: "Invalid params"}
		} else {
			blockNumber := params[0].(string)
			res.Result = sys.eth_getBlockByNumber(blockNumber)
		}

	case "eth_gasPrice":
		res.Result = "0x1"
	case "eth_coinbase":
		res.Result = "1"
	case "eth_blockNumber":
		var params []string
		err := json.Unmarshal(req.Params, &params)
		if err != nil {
			panic(err)
		}
		res.Result = sys.eth_blockNumber()
	case "eth_getBlockByHash":
		res.Result = "1"
	case "eth_getBlockTransactionCountByHash":
		res.Result = "1"
	case "eth_getBlockTransactionCountByNumber":
		res.Result = "1"
	case "eth_getUncleCountByBlockHash":
		res.Result = "1"
	case "eth_getUncleCountByBlockNumber":
		res.Result = "1"
	case "eth_getStorageAt":
		res.Result = "1"

	/*************** Filters ***************/
	case "eth_getLogs":
		res.Result = "1"
	case "eth_subscribe":
		res.Result = "1"
	case "eth_unsubscribe":
		res.Result = "1"

	/*************** Misc ***************/
	case "web3_clientVersion":
		res.Result = "1"
	case "eth_protocolVersion":
		res.Result = "1"
	case "eth_syncing":
		res.Result = "1"
	case "eth_mining":
		res.Result = "1"
	case "eth_hashrate":
		res.Result = "1"
	case "web3_sha3":
		res.Result = "1"
	case "eth_getCompilers":
		res.Result = "1"
	case "eth_compileLLL":
		res.Result = "1"
	case "eth_compileSerpent":
		res.Result = "1"

	/*************** Transactions ***************/
	case "eth_sendTransaction":
		// 第一步：解析为 []map[string]interface{}
		var paramMaps []map[string]interface{}
		if err := json.Unmarshal(req.Params, &paramMaps); err != nil {
			res.Error = &RPCError{Code: -32602, Message: "Invalid params"}
			break
		}
		if len(paramMaps) == 0 {
			res.Error = &RPCError{Code: -32602, Message: "Missing transaction object"}
			break
		}
		param := paramMaps[0] // 通常只有一个交易对象
		// 第二步：显式提取字段
		from, _ := param["from"].(string)
		to, _ := param["to"].(string)
		data, _ := param["data"].(string)
		value, _ := param["value"].(string)
		gas, _ := param["gas"].(string)
		// 第三步：传给处理函数（根据你需要的字段）
		res.Result = sys.eth_sendTransaction(from, to, data, value, gas)

	case "eth_sendRawTransaction":
		res.Result = "1"
	case "eth_getTransactionReceipt":
		var params []interface{}
		if err := json.Unmarshal(req.Params, &params); err != nil || len(params) < 1 {
			res.Error = &RPCError{Code: -32602, Message: "Invalid params"}
		} else {
			txhash := params[0].(string)
			res.Result = sys.eth_getTransactionReceipt(txhash)
		}
	case "eth_getCode":
		res.Result = sys.eth_getCode()
	case "eth_call":
		res.Result = sys.eth_call()
	case "eth_estimateGas":
		res.Result = sys.eth_estimateGas()
	case "eth_maxPriorityFeePerGas":
		res.Result = "0x30000000"
	case "eth_getTransactionCount":
		res.Result = "1"
	case "eth_getTransactionByHash":
		var params []interface{}
		if err := json.Unmarshal(req.Params, &params); err != nil || len(params) < 1 {
			res.Error = &RPCError{Code: -32602, Message: "Invalid params"}
		} else {
			txhash := params[0].(string)
			res.Result = sys.eth_getTransactionByHash(txhash)
		}
	case "eth_getTransactionByBlockHashAndIndex":
		res.Result = "1"
	case "eth_getTransactionByBlockNumberAndIndex":
		res.Result = "1"
	case "eth_getExecutionResultFromSimulator":
		res.Result = "1"
	case "eth_getHHLogsForTx":
		res.Result = "1"
	case "eth_getHashFromTagBySimulator":
		res.Result = "1"
	case "eth_registerCallId":
		res.Result = "1"
	case "eth_getStateDb":
		res.Result = "1"
	case "eth_getBlocksData":
		res.Result = "1"

	/*************** EVM ***************/
	case "evm_setAutomine":
		res.Result = "1"
	case "evm_setIntervalMining":
		res.Result = "1"
	case "evm_snapshot":
		res.Result = "1"
	case "evm_revert":
		res.Result = "1"
	case "evm_increaseTime":
		res.Result = "1"
	case "evm_setNextBlockTimestamp":
		res.Result = "1"
	case "evm_setBlockGasLimit":
		res.Result = "1"
	case "evm_mine":
		res.Result = "1"

	/*************** Miner ***************/
	case "miner_start":
		res.Result = "1"
	case "miner_stop":
		res.Result = "1"

	/*************** Net ***************/
	case "net_version":
		res.Result = sys.net_version()
	case "net_listening":
		res.Result = "1"
	case "net_peerCount":
		res.Result = "1"

	default:
		res.Error = &RPCError{
			Code:    -32601,
			Message: "Method not found",
		}
	}

	return res
}
