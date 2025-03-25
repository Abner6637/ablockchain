package system

import (
	"ablockchain/p2p"
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Commander struct {
	node    *p2p.Node
	reader  *bufio.Reader
	running bool
}

func NewCommander(node *p2p.Node) *Commander {
	return &Commander{
		node:    node,
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
		parts := strings.SplitN(input, " ", 2)

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
		case "peers":
			c.printPeers()
		case "exit":
			c.running = false
		case "help":
			c.printHelp()
		default:
			fmt.Println("未知命令")
		}
	}

	fmt.Println("正在关闭节点...")
	c.node.Host.Close()
}

func (c *Commander) handleConnect(addr string) {
	if err := c.node.ConnectToPeer(addr); err != nil {
		fmt.Printf("连接失败: %v\n", err)
	} else {
		fmt.Println("连接成功")
	}
}

func (c *Commander) handleSend(msg string) {
	if len(c.node.Host.Peerstore().Peers()) < 2 {
		fmt.Println("错误: 未连接任何节点")
		return
	}

	// 发送给第一个连接的节点（实际应用可扩展选择机制）
	targetPeer := c.node.Host.Peerstore().Peers()[1]
	if err := c.node.SendMessage(targetPeer, msg); err != nil {
		fmt.Printf("发送失败: %v\n", err)
	}
}

func (c *Commander) handleBroadcast(msg string) {
	if len(c.node.Host.Peerstore().Peers()) < 2 {
		fmt.Println("错误: 未连接任何节点")
		return
	}
	if err := c.node.BroadcastMessage(msg); err != nil {
		fmt.Printf("发送失败: %v\n", err)
	}
}

func (c *Commander) printPeers() {
	peers := c.node.Host.Network().Peers()
	if len(peers) == 0 {
		fmt.Println("当前没有连接的节点")
		return
	}

	fmt.Println("当前连接的节点列表:")
	for _, peer := range peers {
		fmt.Println(peer.String()) // 打印 Peer ID
	}
}

func (c *Commander) printIncomingMessages() {
	c.node.SetMessageHandler(func(msg string) {
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
  exit                 - 退出程序
  help                 - 显示帮助`)
}
