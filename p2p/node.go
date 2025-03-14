package p2p

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type Node struct {
	Host           host.Host
	ID             string // 当前节点的ID
	Peers          map[peer.ID]peer.AddrInfo
	MessageHandler func(string) // 消息接收回调
}

// 创建新的p2p节点并开始监听
func NewNode(listenAddress string) (*Node, error) {
	node, err := libp2p.New(libp2p.ListenAddrStrings(listenAddress))
	if err != nil {
		return nil, err
	}

	n := &Node{
		Host:  node,
		ID:    node.ID().String(),
		Peers: make(map[peer.ID]peer.AddrInfo),
	}

	// 设置协议处理器
	node.SetStreamHandler(ProtocolID, n.handleStream)

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
	nBytes, err := stream.Read(buf)
	if err != nil {
		fmt.Printf("读取流失败: %v\n", err)
		return
	}

	msg := string(buf[:nBytes])
	fmt.Printf("收到来自 %s 的消息: %s\n", stream.Conn().RemotePeer(), msg)

	// 触发消息回调
	if n.MessageHandler != nil {
		n.MessageHandler(msg)
	}
}

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

	// 将连接成功的peer加入到Peers列表中
	n.Peers[peerInfo.ID] = *peerInfo

	fmt.Printf("Successfully connected to peer: %s, %s\n", peerInfo.Addrs, peerInfo.ID)
	return nil
}

// 广播消息到所有连接的Peer
func (n *Node) BroadcastMessage(message string) error {
	for peerID := range n.Peers {
		if err := n.SendMessage(peerID, message); err != nil {
			return fmt.Errorf("广播消息失败到 %s: %v", peerID, err)
		}
	}
	fmt.Println("消息已广播到所有连接的节点")
	return nil
}
