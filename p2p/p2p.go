package p2p

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type Node struct {
	Host        host.Host
	PeerID      string
	PeerAddress string
}

// 创建并启动一个libp2p节点
func StartNode(listenAddress string, peerAddress string) (*Node, error) {
	node, err := libp2p.New(libp2p.ListenAddrStrings(listenAddress))
	if err != nil {
		return nil, err
	}

	fmt.Println("Node started on", listenAddress)

	nodeID := node.ID()
	fmt.Println("Node Peer ID:", nodeID)

	return &Node{
		Host:        node,
		PeerID:      nodeID.String(),
		PeerAddress: peerAddress,
	}, nil
}

// 连接到对方节点
func (n *Node) ConnectToPeer() error {
	// 将字符串地址转换为 multiaddr.Multiaddr 类型
	peerAddr, err := multiaddr.NewMultiaddr(n.PeerAddress)
	if err != nil {
		return fmt.Errorf("failed to parse multiaddr: %v", err)
	}

	// 获取对方节点的 PeerID，这里需要通过 AddrInfo 来正确构建
	peerInfo := peer.AddrInfo{
		ID:    peer.ID(n.PeerID), // 使用已启动节点的 PeerID 作为目标 PeerID
		Addrs: []multiaddr.Multiaddr{peerAddr},
	}

	// 连接到对方节点
	err = n.Host.Connect(context.Background(), peerInfo)
	if err != nil {
		return fmt.Errorf("failed to connect to peer: %v", err)
	}

	fmt.Println("Successfully connected to peer:", n.PeerAddress)
	return nil
}
