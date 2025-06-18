package api

import (
	"ablockchain/core"
	"ablockchain/event"
	"ablockchain/p2p"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/rlp"
)

func (sys *API) eth_chainId() string {
	return sys.bc.ChainID
}

func (sys *API) eth_getBalance(acc string) string {
	accState, ok := sys.bc.GetAccState(acc)
	if !ok {
		fmt.Printf("账户 %s 不存在于状态数据库中\n", acc)
	}
	return strconv.FormatUint(accState.Balance, 10)
}

func (sys *API) eth_estimateGas() string {
	return "0x1"
}

func (sys *API) eth_getTransactionByHash(txhash string) map[string]interface{} {
	// 去掉前缀 "0x"（不区分大小写）
	cleanInput := strings.TrimPrefix(txhash, "0x")
	cleanInput = strings.TrimPrefix(cleanInput, "0X")
	// 解码 hex 字符串为 []byte
	hashBytes, err := hex.DecodeString(cleanInput)
	if err != nil {
		fmt.Println("解码失败:", err)
		return nil
	}
	index, err := sys.bc.QueryTxIndex([]byte(hashBytes))
	if err != nil {
		fmt.Printf("通过hash获取TxIndex失败: %v\n", err)
		return nil
	}
	tx := sys.bc.GetTxByIndex(index)
	blockNum, err := sys.bc.GetBlockNumByHash(index.BlockHash)
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
		"blockHash":        "0x" + hex.EncodeToString(index.BlockHash),
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

func (sys *API) net_version() string {
	return "1337"
}

// =========================-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-==-=-=-=-=-=-=-=-=-=-
// func (sys *System) eth_requestAccounts() string {
// 	return "0x0000000000000000000000000000000000000001"
// }

func (sys *API) eth_accounts() []string {
	return sys.bc.Acclist //append(sys.blockChain.Acclist, sys.blockChain.Contrlist...)
}

func (sys *API) eth_blockNumber() string {
	return sys.bc.CurBlockNum.String()
}

func (sys *API) eth_getBlockByNumber(num string) map[string]interface{} {
	if num == "latest" {
		key, data, err := sys.bc.DB.GetLatest()
		var value core.Block
		err = rlp.DecodeBytes(data, &value)
		if err != nil {
			fmt.Println("RLP decode error:", err)
		}
		result := map[string]interface{}{
			"number":       "0x" + key,
			"hash":         "0x" + hex.EncodeToString(value.Hash),
			"parentHash":   "0x" + hex.EncodeToString(value.Header.ParentHash),
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

func (sys *API) eth_getCode(addr string) string {
	accState, exists := sys.bc.GetAccState(addr)
	if !exists {
		fmt.Printf("账户 %s 不存在于状态数据库中\n", addr)
	}
	println("Address: " + accState.Address)
	return "0x" + hex.EncodeToString(accState.Code)
}

func (sys *API) eth_call(from string, to string, data string, value string) string {
	storeInput, err := hex.DecodeString(strings.TrimPrefix(data, "0x"))
	if err != nil {
		fmt.Println("数据转换失败:", err)
	}
	result, _, err := sys.bc.RunEVMCall(
		from,
		to,
		storeInput,
		1000000,
		big.NewInt(0),
	)
	if err != nil {
		fmt.Println("合约执行失败:", err)
	}
	fmt.Println("向remix返回结果：")
	fmt.Println("0x" + hex.EncodeToString(result))
	return "0x" + hex.EncodeToString(result)
}

// 部署合约：创建并发布合约交易
func (sys *API) eth_sendTransaction(from string, to string, data string, value string, gas string) string {
	acc, ok := sys.bc.GetAccState(from)
	if !ok {
		fmt.Printf("错误: 付款账户不存在:%s\n", from)
	}
	v, err := strconv.ParseUint(value, 0, 64)
	if err != nil {
		fmt.Println("转换失败:", err)
	}
	if acc.Balance < v {
		fmt.Println("错误: 付款账户余额不足")
	}
	sys.bc.Noncelist[from] += 1 //本地记录nonce值
	if to == "" {
		fmt.Printf("~~~~~~~~~~~~部署合约~~~~~~~~~~~~~~~\n")
		tx := core.NewTransaction(from, "0x", v, data, sys.bc.Noncelist[from])
		signtx, err := core.SignTx(tx) // 本地没有账户私钥
		if err != nil {
			sys.bc.Noncelist[from] -= 1
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
		if err := sys.node.BroadcastMessage(string(encodedP2PMsg)); err != nil {
			fmt.Printf("发送失败: %v\n", err)
		}
		result := "0x" + hex.EncodeToString(tx.TxHash)
		// 添加临时收据所需数据
		sys.tempTxReceipt = core.CalculateContractAddress(from, sys.bc.Noncelist[from])
		return result
	} else {
		fmt.Printf("~~~~~~~~~~~~调用合约~~~~~~~~~~~~~~~\n")
		tx := core.NewTransaction(from, to, v, data, sys.bc.Noncelist[from])
		signtx, err := core.SignTx(tx) // 本地没有账户私钥
		if err != nil {
			sys.bc.Noncelist[from] -= 1
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
		if err := sys.node.BroadcastMessage(string(encodedP2PMsg)); err != nil {
			fmt.Printf("发送失败: %v\n", err)
		}
		result := "0x" + hex.EncodeToString(tx.TxHash)
		// 添加临时收据所需数据（调用合约时直接保存合约地址）
		sys.tempTxReceipt = to

		// 在共识完成，确认交易后，再用evm运行合约
		// storeInput, err := hex.DecodeString(strings.TrimPrefix(data, "0x"))
		// if err != nil {
		// 	fmt.Println("数据转换失败:", err)
		// }
		// // 只进行状态修改，无返回值（要查看状态需要调用eth_call）
		// _, _, err = sys.blockChain.RunEVMCall(
		// 	from,
		// 	to,
		// 	storeInput,
		// 	1000000,
		// 	big.NewInt(0),
		// )
		// if err != nil {
		// 	fmt.Println("合约执行失败:", err)
		// }

		return result
	}

}

func (sys *API) eth_getTransactionReceipt(txhash string) map[string]interface{} {
	// 去掉前缀 "0x"（不区分大小写）
	cleanInput := strings.TrimPrefix(txhash, "0x")
	cleanInput = strings.TrimPrefix(cleanInput, "0X")
	// 解码 hex 字符串为 []byte
	hashBytes, err := hex.DecodeString(cleanInput)
	if err != nil {
		fmt.Println("解码失败:", err)
		return nil
	}

	index, err := sys.bc.QueryTxIndex(hashBytes)
	if err != nil {
		fmt.Printf("交易还未完成共识, 返回预设收据: %v\n", err)
		// 此处需要将合约地址提前返回，如果返回nil，则remix端会不停重复发送eth_getTransactionReceipt
		result := map[string]interface{}{
			"transactionHash":   txhash,
			"contractAddress":   sys.tempTxReceipt,
			"status":            "0x1",
			"blockNumber":       sys.bc.CurBlockNum.String(),
			"cumulativeGasUsed": "0x1",
		}
		return result
	}
	tx := sys.bc.GetTxByIndex(index)
	blockNum, err := sys.bc.GetBlockNumByHash(index.BlockHash)
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

	result := map[string]interface{}{
		"transactionHash":   txhash,
		"contractAddress":   core.CalculateContractAddress(tx.From, tx.Nonce),
		"status":            "0x1",
		"blockNumber":       hexStr,
		"cumulativeGasUsed": "0x1",
	}
	return result
}
