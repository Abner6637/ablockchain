package rpc

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

func Handler(w http.ResponseWriter, r *http.Request) {
	// 处理 CORS
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

	var req RPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("无法解析 JSON-RPC 请求: %s\n原始请求: %s", err, string(body))
		http.Error(w, "Invalid JSON-RPC", http.StatusBadRequest)
		return
	}
	fmt.Printf("**************************\n")
	fmt.Printf("Received method: %s\n", req.Method)
	fmt.Printf("Params: %s\n", string(req.Params))
	fmt.Printf("ID: %s\n", string(*req.ID))
	fmt.Printf("**************************\n")

	res := RPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	/***************
	    Accounts
	****************/
	case "eth_requestAccounts":
		res.Result = eth_requestAccounts()
	case "eth_accounts":
		res.Result = []string{"0x0000000000000000000000000000000000000001"}
	case "eth_getBalance":
		res.Result = "1"
	case "eth_sign":
		res.Result = "1"
	case "eth_chainId":
		res.Result = "1"
	case "eth_signTypedData":
		res.Result = "1"
	case "eth_signTypedData_v4":
		res.Result = "1"
	case "eth_getPKey":
		res.Result = "1"

	/***************
	    Blocks
	****************/
	case "eth_getBlockByNumber":
		res.Result = map[string]interface{}{
			"number":       "0x1a",
			"hash":         "0xabc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abcd",
			"parentHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
			"transactions": []interface{}{},
		}
	case "eth_gasPrice":
		res.Result = "1"
	case "eth_coinbase":
		res.Result = "1"
	case "eth_blockNumber":
		res.Result = "0x1a"
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

	/***************
	    Filters
	****************/
	case "eth_getLogs":
		res.Result = "1"
	case "eth_subscribe":
		res.Result = "1"
	case "eth_unsubscribe":
		res.Result = "1"

	/***************
	      Misc
	****************/
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

	/***************
	  Transactions
	****************/
	case "eth_sendTransaction":
		// 假装交易成功，返回一个伪造哈希
		res.Result = "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	case "eth_sendRawTransaction":
		res.Result = "1"
	case "eth_getTransactionReceipt":
		// 伪造交易收据
		res.Result = map[string]interface{}{
			"transactionHash":   "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			"contractAddress":   "0x1234567890123456789012345678901234567890",
			"status":            "0x1",
			"blockNumber":       "0x1a",
			"cumulativeGasUsed": "0x5208",
		}
	case "eth_getCode":
		// 根据地址决定返回内容，这里简单模拟
		var params []string
		if err := json.Unmarshal(req.Params, &params); err != nil || len(params) < 1 {
			res.Error = &RPCError{
				Code:    -32602,
				Message: "Invalid params",
			}
			break
		}

		address := params[0]
		if address == "0x0000000000000000000000000000000000000002" {
			// 模拟该地址有合约代码
			res.Result = "0x6080604052348015600f57600080fd5b5060..."
		} else {
			// 否则返回空
			res.Result = "0x"
		}
	case "eth_call":
		// 模拟返回 ABI 编码的字符串："Hello from Go"
		res.Result = "0x48656c6c6f2066726f6d20476f" // "Hello from Go" in hex
	case "eth_estimateGas":
		res.Result = "1"
	case "eth_maxPriorityFeePerGas":
		res.Result = "1"
	case "eth_getTransactionCount":
		res.Result = "1"
	case "eth_getTransactionByHash":
		res.Result = map[string]interface{}{
			"hash":             "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			"nonce":            "0x0",
			"blockHash":        "0xabc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abcd",
			"blockNumber":      "0x1a",
			"transactionIndex": "0x0",
			"from":             "0x0000000000000000000000000000000000000001",
			"to":               "0x0000000000000000000000000000000000000002",
			"value":            "0x0",
			"gas":              "0x5208",
			"gasPrice":         "0x3b9aca00",
			"input":            "0x",
			"v":                "0x1b",
			"r":                "0x5e1d3a76fbf824220e3d9a772c88e0f5f3b10b191d257a7d1f0e6a5e4d3c1c7a",
			"s":                "0x2a352c0a204f84bd6cdaa1f07aa0c12323c4e0b94acdf7fda8e4b2b8b3f4b0f5",
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

	/***************
	      Evm
	****************/
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

	/***************
	    Miner
	****************/
	case "miner_start":
		res.Result = "1"
	case "miner_stop":
		res.Result = "1"

	/***************
	      Net
	****************/
	case "net_version":
		res.Result = "1"
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

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("响应编码失败: %s", err)
	}
}

// func Handler(w http.ResponseWriter, r *http.Request) {
// 	// 处理 CORS
// 	w.Header().Set("Access-Control-Allow-Origin", "*")
// 	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
// 	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

// 	if r.Method == http.MethodOptions {
// 		w.WriteHeader(http.StatusOK)
// 		return
// 	}

// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	body, err := io.ReadAll(r.Body)
// 	if err != nil || len(body) == 0 {
// 		http.Error(w, "Invalid or empty body", http.StatusBadRequest)
// 		return
// 	}

// 	var req RPCRequest
// 	if err := json.Unmarshal(body, &req); err != nil {
// 		log.Printf("无法解析 JSON-RPC 请求: %s\n原始请求: %s", err, string(body))
// 		http.Error(w, "Invalid JSON-RPC", http.StatusBadRequest)
// 		return
// 	}
// 	fmt.Printf("**************************\n")
// 	fmt.Printf("Received method: %s\n", req.Method)
// 	fmt.Printf("Params: %s\n", string(req.Params))
// 	fmt.Printf("ID: %s\n", string(*req.ID))
// 	fmt.Printf("**************************\n")

// 	res := RPCResponse{
// 		JSONRPC: "2.0",
// 		ID:      req.ID,
// 	}

// 	switch req.Method {
// 	case "eth_chainId":
// 		res.Result = "1"

// 	case "eth_getBalance":
// 		res.Result = "10000"

// 	case "eth_estimateGas":
// 		res.Result = "100"

// 	case "eth_getTransactionByHash":
// 		res.Result = map[string]interface{}{
// 			"hash":             "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
// 			"nonce":            "0x0",
// 			"blockHash":        "0xabc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abcd",
// 			"blockNumber":      "0x1a",
// 			"transactionIndex": "0x0",
// 			"from":             "0x0000000000000000000000000000000000000001",
// 			"to":               "0x0000000000000000000000000000000000000002",
// 			"value":            "0x0",
// 			"gas":              "0x5208",
// 			"gasPrice":         "0x3b9aca00",
// 			"input":            "0x",
// 			"v":                "0x1b",
// 			"r":                "0x5e1d3a76fbf824220e3d9a772c88e0f5f3b10b191d257a7d1f0e6a5e4d3c1c7a",
// 			"s":                "0x2a352c0a204f84bd6cdaa1f07aa0c12323c4e0b94acdf7fda8e4b2b8b3f4b0f5",
// 		}

// 	case "net_version":
// 		res.Result = "1337"

// 	case "eth_accounts":
// 		res.Result = []string{"0x0000000000000000000000000000000000000001"}

// 	case "eth_blockNumber":
// 		res.Result = "0x1a"

// 	case "eth_getBlockByNumber":
// 		res.Result = map[string]interface{}{
// 			"number":       "0x1a",
// 			"hash":         "0xabc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abcd",
// 			"parentHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
// 			"transactions": []interface{}{},
// 		}

// 	case "eth_getCode":
// 		// 根据地址决定返回内容，这里简单模拟
// 		var params []string
// 		if err := json.Unmarshal(req.Params, &params); err != nil || len(params) < 1 {
// 			res.Error = &RPCError{
// 				Code:    -32602,
// 				Message: "Invalid params",
// 			}
// 			break
// 		}

// 		address := params[0]
// 		if address == "0x0000000000000000000000000000000000000002" {
// 			// 模拟该地址有合约代码
// 			res.Result = "0x6080604052348015600f57600080fd5b5060..."
// 		} else {
// 			// 否则返回空
// 			res.Result = "0x"
// 		}

// 	case "eth_call":
// 		// 模拟返回 ABI 编码的字符串："Hello from Go"
// 		res.Result = "0x48656c6c6f2066726f6d20476f" // "Hello from Go" in hex

// 	case "eth_sendTransaction":
// 		// 假装交易成功，返回一个伪造哈希
// 		res.Result = "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"

// 	case "eth_getTransactionReceipt":
// 		// 伪造交易收据
// 		res.Result = map[string]interface{}{
// 			"transactionHash":   "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
// 			"contractAddress":   "0x1234567890123456789012345678901234567890",
// 			"status":            "0x1",
// 			"blockNumber":       "0x1a",
// 			"cumulativeGasUsed": "0x5208",
// 		}

// 	case "my_custom_method":
// 		// 举例：对自定义方法做回应
// 		// 你可以把 Params 再次 Unmarshal 成具体类型，做更复杂的逻辑
// 		res.Result = map[string]string{
// 			"echo": "你好，Remix！这是 Go 模块的自定义返回。",
// 		}
// 	default:
// 		res.Error = &RPCError{
// 			Code:    -32601,
// 			Message: "Method not found",
// 		}
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	if err := json.NewEncoder(w).Encode(res); err != nil {
// 		log.Printf("响应编码失败: %s", err)
// 	}
// }
