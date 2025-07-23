package system

import (
	"ablockchain/consensus/poa" // 确保正确的导入路径
	"ablockchain/core"
	"ablockchain/event"
	"ablockchain/p2p"
	"bufio"
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
)

type Commander struct {
	sys     *System
	reader  *bufio.Reader
	running bool
	poa     *poacore.Core    // PoA 共识模块
	core    *core.Blockchain // 区块链核心
}

func NewCommander(sys *System) *Commander {
	return &Commander{
		sys:     sys,
		reader:  bufio.NewReader(os.Stdin),
		running: true,
	}
}

func (c *Commander) Run() {
	fmt.Println("输入 'help' 查看可用命令")

	// 启动消息接收协程
	//go c.printIncomingMessages()

	for c.running {
		fmt.Print("> ")
		input, _ := c.reader.ReadString('\n')
		input = strings.TrimSpace(input)
		//parts := strings.SplitN(input, " ", 2)
		parts := strings.Fields(input)

		if len(parts) == 0 {
			fmt.Println("请输入命令，或输入 'help' 查看可用命令")
			continue
		}
		switch parts[0] {
		case "connect":
			if len(parts) < 2 {
				fmt.Println("用法: connect <multiaddr>")
				continue
			}
			c.handleConnect(parts[1])
		case "send":
			if len(parts) < 2 {
				fmt.Println("用法: send <message>")
				continue
			}
			c.handleSend(parts[1])
		case "broadcast":
			if len(parts) < 2 {
				fmt.Println("用法: broadcast <message>")
				continue
			}
			c.handleBroadcast(parts[1])
		case "addval":
			if len(parts) < 3 {
				fmt.Println("用法: addval <address> <value>")
				continue
			}
			value, err := strconv.ParseUint(parts[2], 10, 64) //字符串转为uint64类型, 10:十进制; 64:uint64
			if err != nil {
				fmt.Println("错误: Value 必须是有效的正整数")
				continue
			}
			for i := 1; i <= core.MinTransactionsPerBlock; i++ {
				c.handleTx("0x0000000000000000000000000000000000000001", parts[1], value/uint64(core.MinTransactionsPerBlock))
			}
		case "newacc":
			c.sys.blockChain.NewAccount()
		case "accls":
			c.sys.blockChain.PrintAllAccounts()
		case "miner":
			if len(parts) < 2 {
				fmt.Println("用法: miner <miner address>")
				continue
			}
			c.sys.blockChain.MinerAddr = parts[1]
			c.sys.blockChain.StartMiner()
			fmt.Println("设置矿工账户地址：", parts[1])
		case "tx":
			if len(parts) < 4 {
				fmt.Println("用法: tx <From(adress)> <To(adress)> <Value>")
				continue
			}
			value, err := strconv.ParseUint(parts[3], 10, 64) //字符串转为uint64类型, 10:十进制; 64:uint64
			if err != nil {
				fmt.Println("错误: Value 必须是有效的正整数")
				continue
			}
			c.handleTx(parts[1], parts[2], value)
		case "tx100":
			if len(parts) < 2 {
				fmt.Println("用法: tx100 <From(adress)>")
				continue
			}
			for i := 1; i <= 100; i++ {
				hexStr := fmt.Sprintf("0x%d", i)
				c.handleTx(parts[1], hexStr, uint64(i))
			}
		case "printp":
			c.sys.blockChain.TxPool.PrintPending()
		case "printq":
			c.sys.blockChain.TxPool.PrintQueued()
		case "printlatest":
			c.sys.blockChain.PrintLatest()
		case "printall":
			c.sys.blockChain.PrintAll()
		case "height":
			c.sys.blockChain.DB.PrintCount()
		case "startcons":
			c.sys.consensus.Start()
		case "stopcons":
			c.sys.consensus.Stop()
		case "syc":
			if len(parts) < 2 {
				fmt.Println("用法: syc <addr>")
				continue
			}
			c.sys.Synchronization(parts[1])
		case "testmine":
			if len(parts) < 2 {
				fmt.Println("用法: testmine <blockNumber>")
				continue
			}
			n := new(big.Int)
			n, _ = n.SetString(parts[1], 10)
			c.testmine(n)
		case "peers":
			c.sys.p2pNode.PrintPeers()
		case "id":
			fmt.Printf("listenAddr: %v\n", c.sys.p2pNode.Host.Addrs())
			fmt.Printf("p2pID:%s", c.sys.p2pNode.ID)
		case "valadd":
			if len(parts) < 2 {
				fmt.Println("用法: valadd <address>")
				return
			}
			if poaCore, ok := c.sys.consensus.(*poacore.Core); ok {
				poaCore.AddValidator(parts[1]) // 直接调用 Core 方法
			} else {
				fmt.Println("当前共识模块不支持验证者管理操作")
			}

		case "reval":
			if len(parts) < 2 {
				fmt.Println("用法: reval <address>")
				return
			}
			if poaCore, ok := c.sys.consensus.(*poacore.Core); ok {
				poaCore.RemoveValidator(parts[1]) // 直接调用 Core 方法
			} else {
				fmt.Println("当前共识模块不支持验证者管理操作")
			}

		case "listvals":
			if poaCore, ok := c.sys.consensus.(*poacore.Core); ok {
				poaCore.ListValidators() // 直接调用 Core 方法
			} else {
				fmt.Println("当前共识模块不支持验证者管理操作")
			}

		case "exit":
			c.running = false
		case "help":
			c.printHelp()
		default:
			fmt.Println("未知命令")
		}
	}

	fmt.Println("正在关闭节点...")
	c.sys.p2pNode.Host.Close()

	event.Bus.Publish("ExitEvent", true)
	c.sys.consensus.Stop()
}

func (c *Commander) handleConnect(addr string) {
	if err := c.sys.p2pNode.ConnectToPeer(addr); err != nil {
		fmt.Printf("连接失败: %v\n", err)
	} else {
		c.sys.Synchronization(strings.Split(addr, "/p2p/")[1])
		fmt.Println("连接成功, 进行同步...")
	}
}

func (c *Commander) handleSend(msg string) {
	if len(c.sys.p2pNode.Peers()) < 1 {
		fmt.Println("错误: 未连接任何节点")
		return
	}

	// 发送给第一个连接的节点（实际应用可扩展选择机制）
	targetPeer := c.sys.p2pNode.Peers()[0]
	if err := c.sys.p2pNode.SendMessage(targetPeer, msg); err != nil {
		fmt.Printf("发送失败: %v\n", err)
	}
}

func (c *Commander) handleBroadcast(msg string) {
	if len(c.sys.p2pNode.Peers()) < 1 {
		fmt.Println("错误: 未连接任何节点")
		return
	}
	if err := c.sys.p2pNode.BroadcastMessage(msg); err != nil {
		fmt.Printf("发送失败: %v\n", err)
	}
}

func (c *Commander) handleAddVal(address string, value uint64) {
	acc, ok := c.sys.blockChain.GetAccState(address)
	if !ok {
		fmt.Println("错误: 账户不存在")
		return
	}
	acc.Balance += value
	c.sys.blockChain.UpdateAccState(acc)
}

// 发送交易
func (c *Commander) handleTx(from, to string, value uint64) {
	// 优先处理水龙头账户（跳过验证直接生成交易）
	if from == "0x0000000000000000000000000000000000000001" {
		c.sys.blockChain.Noncelist[from] += 1
		tx := core.NewTransaction(from, to, value, "", c.sys.blockChain.Noncelist[from])
		signedTx := &core.SignedTx{
			Tx:   tx,
			Sign: nil,
		}
		data, err := rlp.EncodeToBytes(signedTx)
		if err != nil {
			log.Fatalf("rlp err: %v", err)
		}
		p2pMsg := &p2p.Message{
			Type: p2p.TransactionMessage,
			Data: data,
		}
		encodedP2PMsg, err := p2pMsg.Encode()
		if err != nil {
			log.Fatal("消息编码失败:encodedP2PMsg, err := p2pMsg.Encode()")
		}
		event.Bus.Publish("TransactionMessage", signedTx) //自己节点也要把交易加入交易池,blockchain.StartMiner中处理
		if err := c.sys.p2pNode.BroadcastMessage(string(encodedP2PMsg)); err != nil {
			fmt.Printf("发送失败: %v\n", err)
		}
		return
	}

	// 普通账户
	acc, ok := c.sys.blockChain.GetAccState(from)
	if !ok {
		fmt.Println("错误: 付款账户不存在")
		return
	}
	if acc.Balance < value {
		fmt.Println("错误: 付款账户余额不足")
		return
	}
	c.sys.blockChain.Noncelist[from] += 1 //本地记录nonce值
	tx := core.NewTransaction(from, to, value, "", c.sys.blockChain.Noncelist[from])
	signtx, err := core.SignTx(tx) // 本地没有账户私钥
	if err != nil {
		c.sys.blockChain.Noncelist[from] -= 1
		return
	}
	data, err := rlp.EncodeToBytes(signtx)
	if err != nil {
		log.Fatalf("rlp err: %v", err)
	}
	p2pMsg := &p2p.Message{
		Type: p2p.TransactionMessage,
		Data: data,
	}
	encodedP2PMsg, err := p2pMsg.Encode()
	if err != nil {
		log.Fatal("消息编码失败:encodedP2PMsg, err := p2pMsg.Encode()")
	}
	event.Bus.Publish("TransactionMessage", signtx) //自己节点也要把交易加入交易池,blockchain.StartMiner中处理
	if err := c.sys.p2pNode.BroadcastMessage(string(encodedP2PMsg)); err != nil {
		fmt.Printf("发送失败: %v\n", err)
	} else {
		log.Printf("\n@@TX_SUBMITTED TxHash=%x time=%v\n", tx.TxHash, time.Now())
	}
}

// 暂未使用
func (c *Commander) printIncomingMessages() {
	c.sys.p2pNode.SetMessageHandler(func(msg string, id string) {
		fmt.Printf("\n[新消息] %s\n> ", msg) // 保持输入提示符
	})
}

func (c *Commander) printHelp() {
	fmt.Println(`
可用命令:
  connect <multiaddr>  - 连接到指定节点
  send <message>       - 发送消息
  broadcast <message>  - 广播消息
  peers                - 打印peers节点列表
  newacc               - 创建新账户
  accls                - 打印账户列表
  tx <From(adress)> <To(adress)> <Value> -发送交易
  syc <addr>           - 同步
  printlatest          - 打印最新区块
  printall             - 打印所有区块
  height               - 总区块数
  testmine             - 测试共识
  valadd <address>      - 添加验证者
  reval <address>       - 删除验证者
  listvals              - 列出所有验证者
  exit                 - 退出程序
  help                 - 显示帮助`)
}

func (c *Commander) testmine(num *big.Int) {
	c.sys.blockChain.NewBlockChan <- newTestBlock(num)

}

func newTestBlock(num *big.Int) *core.Block {
	return &core.Block{
		Header: &core.BlockHeader{
			ParentHash: []byte("0df9a8f4a2f2fc354c3c8aa5e837d4db137f20ccbf3d8336e4c95ac9d0e2943e"),
			MerkleRoot: []byte("1cdfdf5680f2a639732f6aae64a8b96c10a913b46c8fcd908c9eb95925979974"),
			Time:       uint64(time.Now().Unix()),
			Difficulty: 2,
			Nonce:      0,
			Number:     num,
		},
	}
}
