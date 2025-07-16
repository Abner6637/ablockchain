package system

import (
	"ablockchain/core"
	"ablockchain/event"
	"ablockchain/p2p"
	"bufio"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"

	poacore "ablockchain/consensus/poa/core"

	"github.com/ethereum/go-ethereum/rlp"
)

type IPCCommander struct {
	sys     *System
	reader  *bufio.Reader
	running bool
	cmdCh   chan string
	input   io.Reader
	output  io.Writer
}

func NewIPCCommander(sys *System, in io.Reader, out io.Writer) *IPCCommander {
	return &IPCCommander{
		sys:     sys,
		reader:  bufio.NewReader(os.Stdin),
		running: true,
		cmdCh:   make(chan string),
		input:   in,
		output:  out,
	}
}

func (c *IPCCommander) Run() {
	c.output.Write([]byte("输入 'help' 查看可用命令\n"))
	c.output.Write([]byte("> "))
	// 启动消息接收协程
	//go c.printIncomingMessages()

	go c.listenForCommands()

	for c.running {
		select {
		case cmd := <-c.cmdCh:
			c.executeCommand(cmd)
			c.output.Write([]byte("> "))
		}
	}
}

func (c *IPCCommander) listenForCommands() {

	scanner := bufio.NewScanner(c.input)
	// 持续监听 attach 进程的输入并将命令传递给 main goroutine
	for scanner.Scan() {

		cmd := scanner.Text()
		input := strings.TrimSpace(cmd)
		if len(input) == 0 {
			c.output.Write([]byte("请输入命令，或输入 'help' 查看可用命令\n"))
			c.output.Write([]byte("> "))
			continue
		}

		// c.conn.Write([]byte("发送数据" + input + "给主进程"))
		c.cmdCh <- input // 将命令传递给主进程
	}
}

func (c *IPCCommander) executeCommand(cmd string) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}
	switch parts[0] {
	case "connect":
		if len(parts) < 2 {
			c.output.Write([]byte("用法: connect <multiaddr>\n"))
			return
		}
		c.handleConnect(parts[1])
	case "send":
		if len(parts) < 2 {
			c.output.Write([]byte("用法: send <message>\n"))
			return
		}
		c.handleSend(parts[1])
	case "broadcast":
		if len(parts) < 2 {
			c.output.Write([]byte("用法: broadcast <message>\n"))
			return
		}
		c.handleBroadcast(parts[1])
	case "addval":
		if len(parts) < 3 {
			fmt.Println("用法: addval <address> <value>")
			return
		}
		value, err := strconv.ParseUint(parts[2], 10, 64) //字符串转为uint64类型, 10:十进制; 64:uint64
		if err != nil {
			fmt.Println("错误: Value 必须是有效的正整数")
			return
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
			c.output.Write([]byte("用法: miner <miner address>\n"))
			return
		}
		c.sys.blockChain.MinerAddr = parts[1]
		c.sys.blockChain.StartMiner()
		fmt.Println("设置矿工账户地址：", parts[1])
	case "tx":
		if len(parts) < 4 {
			c.output.Write([]byte("用法: tx <From(adress)> <To(adress)> <Value>\n"))
			return
		}
		value, err := strconv.ParseUint(parts[3], 10, 64) //字符串转为uint64类型, 10:十进制; 64:uint64
		if err != nil {
			c.output.Write([]byte("错误: Value 必须是有效的正整数\n"))
			return
		}
		c.handleTx(parts[1], parts[2], value)
	case "tx100":
		if len(parts) < 2 {
			fmt.Println("用法: tx100 <From(adress)>")
			return
		}
		for i := 1; i <= 100; i++ {
			hexStr := fmt.Sprintf("0x%d", i)
			c.handleTx(parts[1], hexStr, uint64(i))
		}
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
			return
		}
		c.sys.Synchronization(parts[1])
	case "testmine":
		if len(parts) < 2 {
			c.output.Write([]byte("用法: testmine <blockNumber>\n"))
			return
		}
		n := new(big.Int)
		n, _ = n.SetString(parts[1], 10)
		c.testmine(n)
	case "peers":
		c.PrintPeers()
	case "id":
		str1 := fmt.Sprintf("listenAddr: %v\n", c.sys.p2pNode.Host.Addrs())
		str2 := fmt.Sprintf("p2pID:%s\n", c.sys.p2pNode.ID)
		c.output.Write([]byte(str1))
		c.output.Write([]byte(str2))

	case "valadd":
		if len(parts) < 2 {
			fmt.Println("用法: valadd <address>")
			return
		}
		if poaCore, ok := c.sys.consensus.(*poacore.Core); ok {
			poaCore.AddValidator(parts[1]) // 增加验证者
			fmt.Println("验证者已添加（如地址合法）")
		} else {
			fmt.Println("当前共识模块不支持验证者管理操作")
		}

	case "valdel":
		if len(parts) < 2 {
			fmt.Println("用法: valdel <address>")
			return
		}
		if poaCore, ok := c.sys.consensus.(*poacore.Core); ok {
			poaCore.RemoveValidator(parts[1]) // 删除验证者
			fmt.Println("验证者已删除（如地址合法）")
		} else {
			fmt.Println("当前共识模块不支持验证者管理操作")
		}

	case "vallist":
		if poaCore, ok := c.sys.consensus.(*poacore.Core); ok {
			if len(poaCore.ValSet) == 0 {
				fmt.Println("验证者列表为空")
			} else {
				fmt.Println("当前验证者列表:")
				for _, v := range poaCore.ValSet {
					fmt.Printf("  0x%x\n", v)
				}
			}
		} else {
			fmt.Println("当前共识模块不支持验证者管理操作")
		}

	case "exit":
		c.running = false

		c.output.Write([]byte("正在关闭节点...\n"))
		c.sys.p2pNode.Host.Close()

		event.Bus.Publish("ExitEvent", true)
		// log.Printf("发布ExitEvent事件")

		c.sys.consensus.Stop()

	case "help":
		c.printHelp()
	default:
		c.output.Write([]byte("未知命令\n"))
	}
}

func (c *IPCCommander) handleConnect(addr string) {
	if err := c.sys.p2pNode.ConnectToPeer(addr); err != nil {
		str := fmt.Sprintf("连接失败: %v\n", err)
		c.output.Write([]byte(str))
	} else {
		fmt.Println("连接成功")
	}
}

func (c *IPCCommander) handleSend(msg string) {
	if len(c.sys.p2pNode.Peers()) < 1 {
		c.output.Write([]byte("错误: 未连接任何节点\n"))
		return
	}

	// 发送给第一个连接的节点（实际应用可扩展选择机制）
	targetPeer := c.sys.p2pNode.Peers()[0]
	if err := c.sys.p2pNode.SendMessage(targetPeer, msg); err != nil {
		fmt.Printf("发送失败: %v\n", err)
	}
}

func (c *IPCCommander) handleBroadcast(msg string) {
	if len(c.sys.p2pNode.Peers()) < 1 {
		fmt.Println("错误: 未连接任何节点")
		return
	}
	if err := c.sys.p2pNode.BroadcastMessage(msg); err != nil {
		fmt.Printf("发送失败: %v\n", err)
	}
}

func (c *IPCCommander) handleAddVal(address string, value uint64) {
	acc, ok := c.sys.blockChain.GetAccState(address)
	if !ok {
		fmt.Println("错误: 账户不存在")
		return
	}
	acc.Balance += value
	c.sys.blockChain.UpdateAccState(acc)
}

// 发送交易
func (c *IPCCommander) handleTx(from, to string, value uint64) {
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
		log.Fatalf("SignTx err: %v", err)
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
	}
}

// 暂未使用
func (c *IPCCommander) printIncomingMessages() {
	c.sys.p2pNode.SetMessageHandler(func(msg string, id string) {
		fmt.Printf("\n[新消息] %s\n> ", msg) // 保持输入提示符
	})
}

func (c *IPCCommander) printHelp() {
	c.output.Write([]byte(`
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
  valadd <address>     - 添加验证者地址
  valdel <address>     - 删除验证者地址
  vallist              - 查看验证者列表
  exit                 - 退出程序
  help                 - 显示帮助` + "\n"))
}

func (c *IPCCommander) PrintPeers() {
	peers := c.sys.p2pNode.Peers() // 获取当前连接的所有对等节点

	if len(peers) == 0 {
		c.output.Write([]byte("当前没有连接的对等节点\n"))
		return
	}

	c.output.Write([]byte("当前连接的对等节点:\n"))
	for _, peerID := range peers {
		str := fmt.Sprintf("- %v\n", peerID)
		c.output.Write([]byte(str))
	}
}

func (c *IPCCommander) testmine(num *big.Int) {
	c.sys.blockChain.NewBlockChan <- newTestBlock(num)

}
