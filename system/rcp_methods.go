package system

import (
	"ablockchain/core"
	"ablockchain/event"
	"ablockchain/p2p"
	"fmt"
	"log"
	"strconv"

	"github.com/ethereum/go-ethereum/rlp"
)

func (sys *System) eth_chainId() string {
	return sys.blockChain.ChainID
}

func (sys *System) eth_getBalance(acc string) string {
	accState, ok := sys.blockChain.GetAccState(acc)
	if !ok {
		fmt.Printf("账户 %s 不存在于状态数据库中\n", acc)
	}
	return strconv.FormatUint(accState.Balance, 10)
}

func (sys *System) eth_estimateGas() string {
	return "100"
}

func (sys *System) eth_getTransactionByHash(txhash string) map[string]interface{} {
	index, err := sys.blockChain.QueryTxIndex([]byte(txhash))
	if err != nil {
		fmt.Printf("通过hash获取TxIndex失败: %v\n", err)
		return nil
	}
	tx := sys.blockChain.GetTxByIndex(index)
	blockNum, err := sys.blockChain.GetBlockNumByHash(index.BlockHash)
	if err != nil {
		fmt.Printf("通过hash获取区块失败: %v\n", err)
		return nil
	}
	//将字符串转换为整数
	num, err := strconv.Atoi(blockNum)
	if err != nil {
		fmt.Println("转换错误:", err)
		return nil
	}
	// 将整数格式化为带有0x前缀的十六进制字符串
	hexStr := fmt.Sprintf("0x%X", num)

	//  v,r,s为ECDSA 签名参数（用于验证交易签名），构成交易的签名部分。
	//  - v: 恢复参数，表明使用哪个公钥恢复出签名者地址。
	//	- r 和 s: 椭圆曲线签名的两个部分。
	result := map[string]interface{}{
		"hash":             txhash,
		"nonce":            fmt.Sprintf("0x%x", tx.Nonce),
		"blockHash":        string(index.BlockHash),
		"blockNumber":      hexStr,
		"transactionIndex": fmt.Sprintf("0x%x", index.TxIndex),
		"from":             tx.From,
		"to":               tx.To,
		"value":            fmt.Sprintf("0x%x", tx.Value),
		"gas":              "0x5208",
		"gasPrice":         "0x3b9aca00",
		"input":            tx.Data,
		"v":                "0x1b",
		"r":                "0x5e1d3a76fbf824220e3d9a772c88e0f5f3b10b191d257a7d1f0e6a5e4d3c1c7a",
		"s":                "0x2a352c0a204f84bd6cdaa1f07aa0c12323c4e0b94acdf7fda8e4b2b8b3f4b0f5",
	}
	fmt.Printf("%+v\n", result)
	return result
}

func (sys *System) net_version() string {
	return "1337"
}

// =========================-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-==-=-=-=-=-=-=-=-=-=-
func (sys *System) eth_requestAccounts() string {
	return "0x0000000000000000000000000000000000000001"
}

func (sys *System) eth_accounts() []string {
	return append(sys.blockChain.Acclist, sys.blockChain.Contrlist...)
}

func (sys *System) eth_blockNumber() string {
	return "0x1a"
}

func (sys *System) eth_getBlockByNumber(num string) map[string]interface{} {
	if num == "latest" {
		key, data, err := sys.blockChain.DB.GetLatest()
		var value core.Block
		err = rlp.DecodeBytes(data, &value)
		if err != nil {
			fmt.Println("RLP decode error:", err)
		}
		result := map[string]interface{}{
			"number":       key,
			"hash":         value.Hash,
			"parentHash":   value.Header.ParentHash,
			"transactions": value.Transactions,
		}
		return result
	} else {
		result := map[string]interface{}{
			"number":       "0x1a",
			"hash":         "0xabc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abcd",
			"parentHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
			"transactions": []interface{}{},
		}
		return result
	}
}

func (sys *System) eth_getCode() string {
	// 根据地址决定返回内容，这里简单模拟
	// var params []string
	// if err := json.Unmarshal(req.Params, &params); err != nil || len(params) < 1 {
	// 	res.Error = &RPCError{
	// 		Code:    -32602,
	// 		Message: "Invalid params",
	// 	}
	// 	break
	// }

	// address := params[0]
	// if address == "0x0000000000000000000000000000000000000002" {
	// 	// 模拟该地址有合约代码
	// 	res.Result = "0x6080604052348015600f57600080fd5b5060..."
	// } else {
	// 	// 否则返回空
	// 	res.Result = "0x"
	// }
	return "0x6080604052348015600f57600080fd5b5060..."
}

func (sys *System) eth_call() string {
	// 模拟返回 ABI 编码的字符串："Hello from Go"
	result := "0x48656c6c6f2066726f6d20476f"
	return result
}

// 部署合约：创建并发布合约交易
func (sys *System) eth_sendTransaction(from string, data string, value string, gas string) string {
	acc, ok := sys.blockChain.GetAccState(from)
	if !ok {
		fmt.Println("错误: 付款账户不存在")
	}
	v, err := strconv.ParseUint(value, 0, 64)
	if err != nil {
		fmt.Println("转换失败:", err)
	}
	if acc.Balance < v {
		fmt.Println("错误: 付款账户余额不足")
	}
	sys.blockChain.Noncelist[from] += 1 //本地记录nonce值
	tx := core.NewTransaction(from, "0x", v, data, sys.blockChain.Noncelist[from])
	signtx, err := core.SignTx(tx) // 本地没有账户私钥
	if err != nil {
		sys.blockChain.Noncelist[from] -= 1
	}
	p2pdata, err := rlp.EncodeToBytes(signtx)
	if err != nil {
		log.Fatalf("rlp err: %v", err)
	}
	p2pMsg := &p2p.Message{
		Type: p2p.TransactionMessage,
		Data: p2pdata,
	}
	encodedP2PMsg, err := p2pMsg.Encode()
	if err != nil {
		log.Fatal("消息编码失败:encodedP2PMsg, err := p2pMsg.Encode()")
	}
	event.Bus.Publish("TransactionMessage", signtx) //自己节点也要把交易加入交易池,blockchain.StartMiner中处理
	if err := sys.p2pNode.BroadcastMessage(string(encodedP2PMsg)); err != nil {
		fmt.Printf("发送失败: %v\n", err)
	}
	result := tx.TxHash
	return string(result)
}

func (sys *System) eth_getTransactionReceipt(txhash string) map[string]interface{} {
	index, err := sys.blockChain.QueryTxIndex([]byte(txhash))
	if err != nil {
		fmt.Printf("通过hash获取TxIndex失败: %v\n", err)
		return nil
	}
	tx := sys.blockChain.GetTxByIndex(index)
	blockNum, err := sys.blockChain.GetBlockNumByHash(index.BlockHash)
	if err != nil {
		fmt.Printf("通过hash获取区块失败: %v\n", err)
		return nil
	}
	//将字符串转换为整数
	num, err := strconv.Atoi(blockNum)
	if err != nil {
		fmt.Println("转换错误:", err)
		return nil
	}
	// 将整数格式化为带有0x前缀的十六进制字符串
	hexStr := fmt.Sprintf("0x%X", num)

	// 伪造交易收据
	result := map[string]interface{}{
		"transactionHash":   string(tx.TxHash),
		"contractAddress":   core.CalculateContractAddress(tx.From, tx.Nonce),
		"status":            "0x1",
		"blockNumber":       hexStr,
		"cumulativeGasUsed": "0x5208",
	}
	return result
}
