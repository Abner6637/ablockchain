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
	PeerID      string // 当前节点的 PeerID
	PeerAddress string // 对方节点的完整地址（需包含其 PeerID）
}

func StartNode(listenAddress string) (*Node, error) {
	node, err := libp2p.New(libp2p.ListenAddrStrings(listenAddress))
	if err != nil {
		return nil, err
	}

	fmt.Printf("Node started on %s, PeerID: %s\n", listenAddress, node.ID())
	return &Node{
		Host:   node,
		PeerID: node.ID().String(),
	}, nil
}

func (n *Node) ConnectToPeer(peerAddress string) error {
	// 解析对方节点的完整地址（需包含 PeerID）
	peerAddr, err := multiaddr.NewMultiaddr(peerAddress)
	if err != nil {
		return fmt.Errorf("failed to parse peer address: %v", err)
	}

	// 从 multiaddr 中提取对方节点的 PeerID 和地址
	peerInfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
	if err != nil {
		return fmt.Errorf("failed to extract peer info: %v", err)
	}

	// 连接到对方节点
	if err := n.Host.Connect(context.Background(), *peerInfo); err != nil {
		return fmt.Errorf("failed to connect to peer: %v", err)
	}

	fmt.Printf("Successfully connected to peer: %s\n", peerInfo.ID)
	return nil
}
