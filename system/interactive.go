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
	"time"

	"github.com/ethereum/go-ethereum/rlp"
)

type Commander struct {
	sys     *System
	reader  *bufio.Reader
	running bool
	cmdCh   chan string
	input   io.Reader
	output  io.Writer
}

func NewCommander(sys *System, in io.Reader, out io.Writer) *Commander {
	return &Commander{
		sys:     sys,
		reader:  bufio.NewReader(os.Stdin),
		running: true,
		cmdCh:   make(chan string),
		input:   in,
		output:  out,
	}
}

func (c *Commander) Run() {
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

func (c *Commander) listenForCommands() {

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

func (c *Commander) executeCommand(cmd string) {
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
			c.output.Write([]byte("用法: addval <address> <value>\n"))
			return
		}
		value, err := strconv.ParseUint(parts[2], 10, 64) //字符串转为uint64类型, 10:十进制; 64:uint64
		if err != nil {
			c.output.Write([]byte("错误: Value 必须是有效的正整数\n"))
			return
		}
		c.handleAddVal(parts[1], value)
	case "newacc":
		c.sys.blockChain.StateDB.NewAccount()
	case "accls":
		c.sys.blockChain.StateDB.PrintAccounts()
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
	case "testmine":
		c.testmine()
	case "peers":
		c.PrintPeers()
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

func (c *Commander) handleConnect(addr string) {
	if err := c.sys.p2pNode.ConnectToPeer(addr); err != nil {
		str := fmt.Sprintf("连接失败: %v\n", err)
		c.output.Write([]byte(str))
	} else {
		fmt.Println("连接成功")
	}
}

func (c *Commander) handleSend(msg string) {
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
	acc, ok := c.sys.blockChain.StateDB.GetAccount(address)
	if !ok {
		fmt.Println("错误: 账户不存在")
		return
	}
	acc.Balance += value
	c.sys.blockChain.StateDB.UpdateAccount(acc)
}

// 发送交易
func (c *Commander) handleTx(from, to string, value uint64) {
	acc, ok := c.sys.blockChain.StateDB.GetAccount(from)
	if !ok {
		fmt.Println("错误: 付款账户不存在")
		return
	}
	if acc.Balance < value {
		fmt.Println("错误: 付款账户余额不足")
		return
	}
	acc.Nonce += 1 //Nonce在这里增加：防止一个区块中有多个相同用户的交易导致错误(不正规)
	c.sys.blockChain.StateDB.UpdateAccount(acc)
	tx := core.NewTransaction(acc, to, value)
	signtx, err := acc.SignTx(tx)
	data, err := rlp.EncodeToBytes(signtx)
	if err != nil {
		log.Fatalf("SignTx err", err)
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
func (c *Commander) printIncomingMessages() {
	c.sys.p2pNode.SetMessageHandler(func(msg string) {
		fmt.Printf("\n[新消息] %s\n> ", msg) // 保持输入提示符
	})
}

func (c *Commander) printHelp() {
	c.output.Write([]byte(`
可用命令:
  connect <multiaddr>  - 连接到指定节点
  send <message>       - 发送消息
  broadcast <message>  - 广播消息
  peers                - 打印peers节点列表
  newacc               - 创建新账户
  accls                - 打印账户列表
  tx <From(adress)> <To(adress)> <Value> -发送交易
  printlatest          - 打印最新区块
  printall             - 打印所有区块
  height               - 总区块数
  testmine             - 测试共识
  exit                 - 退出程序
  help                 - 显示帮助` + "\n"))
}

func (c *Commander) PrintPeers() {
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

func (c *Commander) testmine() {
	c.sys.blockChain.NewBlockChan <- newTestBlock()

}

func newTestBlock() *core.Block {
	return &core.Block{
		Header: &core.BlockHeader{
			ParentHash: []byte("0df9a8f4a2f2fc354c3c8aa5e837d4db137f20ccbf3d8336e4c95ac9d0e2943e"),
			MerkleRoot: []byte("1cdfdf5680f2a639732f6aae64a8b96c10a913b46c8fcd908c9eb95925979974"),
			Time:       uint64(time.Now().Unix()),
			Difficulty: 2,
			Nonce:      0,
			Number:     big.NewInt(13),
		},
	}
}
