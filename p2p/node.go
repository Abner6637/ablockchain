package p2p

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type Node struct {
	Host           host.Host
	ID             string       // 当前节点的ID
	MessageHandler func(string) // 消息接收回调
}

// 创建新的p2p节点并开始监听
func NewNode(listenAddress string) (*Node, error) {
	node, err := libp2p.New(libp2p.ListenAddrStrings(listenAddress))
	if err != nil {
		return nil, err
	}

	n := &Node{
		Host: node,
		ID:   node.ID().String(),
	}

	// 设置协议处理器
	node.SetStreamHandler(ProtocolID, n.handleStream)

	// handleStream中消息回调时，触发MessageHandler，处理消息
	// TODO：p2p层处理消息时是否需要另外建立一个进程？
	n.SetMessageHandler(func(msg string) {
		fmt.Printf("\n[消息] %s\n> ", msg)
		decodedMsg, err := DecodeMessage([]byte(msg))
		if err != nil {
			fmt.Printf("RLP解码失败\n")
			return
		}
		ProcessMessage(decodedMsg)
		log.Println("message handler处理消息完毕")
	})

	fmt.Printf("Node started on %s, ID: %s\n", node.Addrs(), node.ID())
	return n, nil
}

// 设置消息处理器
func (n *Node) SetMessageHandler(handler func(string)) {
	n.MessageHandler = handler
}

func (n *Node) handleStream(stream network.Stream) {
	defer stream.Close()

	// 读取对方发送的数据
	buf := make([]byte, 1024)
	var fullData []byte // 定义一个fullData，防止消息长度大于1024字节时无法完全读取
	for {
		n, err := stream.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("读取数据时出错: %s", err)
		}
		fullData = append(fullData, buf[:n]...)
	}

	//读取消息源节点的信息
	peerID := stream.Conn().RemotePeer()
	msg := string(fullData)
	fmt.Printf("收到来自 %s 的消息: %s\n", peerID, msg)

	// 触发消息回调
	if n.MessageHandler != nil {
		n.MessageHandler(msg)
	}
}

// 发送消息，需要增加消息的类型
//
// TODO
func (n *Node) SendMessage(peerID peer.ID, message string) error {
	// 创建流
	stream, err := n.Host.NewStream(context.Background(), peerID, ProtocolID)
	if err != nil {
		return fmt.Errorf("创建流失败: %v", err)
	}
	defer stream.Close()

	// 写入数据
	if _, err = stream.Write([]byte(message)); err != nil {
		return fmt.Errorf("写入消息失败: %v", err)
	}

	fmt.Printf("已发送消息到 %s: %s\n", peerID, message)
	return nil
}

func (n *Node) ConnectToPeer(peerAddress string) error {
	// 将string类型的地址转换为multiaddr类型
	peerAddr, err := multiaddr.NewMultiaddr(peerAddress)
	if err != nil {
		return fmt.Errorf("failed to parse peer address: %v", err)
	}

	// 从peerAddr中提取peer节点的Info，该Info包含地址和ID
	peerInfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
	if err != nil {
		return fmt.Errorf("failed to extract peer info: %v", err)
	}

	// 连接到peer节点
	if err := n.Host.Connect(context.Background(), *peerInfo); err != nil {
		return fmt.Errorf("failed to connect to peer: %v", err)
	}

	fmt.Printf("Successfully connected to peer: %s, %s\n", peerInfo.Addrs, peerInfo.ID)
	return nil
}

// Peers 方法动态获取当前连接的所有节点
func (n *Node) Peers() []peer.ID {
	return n.Host.Network().Peers()
}

func (n *Node) BroadcastMessage(message string) error {
	for _, peerID := range n.Peers() {
		if err := n.SendMessage(peerID, message); err != nil {
			return fmt.Errorf("广播消息失败到 %s: %v", peerID, err)
		}
	}
	fmt.Println("消息已广播到所有连接的节点")
	return nil
}

func (n *Node) PrintPeers() {
	peers := n.Peers() // 获取当前连接的所有对等节点

	if len(peers) == 0 {
		fmt.Println("当前没有连接的对等节点")
		return
	}

	fmt.Println("当前连接的对等节点:")
	for _, peerID := range peers {
		fmt.Println("- ", peerID)
	}
}
