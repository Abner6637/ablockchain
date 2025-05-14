package p2p

import (
	"ablockchain/crypto"
	"context"
	"crypto/ecdsa"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/libp2p/go-libp2p"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type Node struct {
	Host           host.Host
	ID             string               // 当前节点的ID
	MessageHandler func(string, string) // 消息接收回调

	PrivateKey *ecdsa.PrivateKey
}

// 创建新的p2p节点并开始监听
func NewNode(listenAddress string, privateKey *ecdsa.PrivateKey, libPriv crypto.LibPrivKey) (*Node, error) {

	node, err := libp2p.New(libp2p.ListenAddrStrings(listenAddress),
		libp2p.Identity(libPriv),
		libp2p.NoSecurity,
	)

	if err != nil {
		return nil, err
	}

	n := &Node{
		Host:       node,
		ID:         node.ID().String(),
		PrivateKey: privateKey,
	}

	// 启动两次MDNS，尽量连通吧……
	// TODO: 理应来讲应该不需要使用goroutine，保证连通了之后，再开启共识等
	// 但是会和mesh.sh冲突
	// go setupMDNS(n.Host, "my-libp2p-network")
	// go setupMDNS(n.Host, "my-libp2p-network2")

	// 设置协议处理器
	node.SetStreamHandler(ProtocolID, n.handleStream)

	// handleStream中消息回调时，触发MessageHandler，处理消息
	// TODO：p2p层处理消息时是否需要另外建立一个进程？
	n.SetMessageHandler(func(msg string, id string) {
		// fmt.Printf("\np2p处理[消息] %s\n> ", msg)
		decodedMsg, err := DecodeMessage([]byte(msg))
		if err != nil {
			fmt.Printf("RLP解码失败\n")
			return
		}
		ProcessMessage(decodedMsg, id)
		// log.Println("p2p处理消息完毕")
	})

	fmt.Printf("Node started on %s, ID: %s\n", node.Addrs(), node.ID())
	return n, nil
}

// 设置消息处理器
func (n *Node) SetMessageHandler(handler func(string, string)) {
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
	fmt.Printf("\np2p收到来自 %s 的消息\n", peerID)

	// 触发消息回调
	if n.MessageHandler != nil {
		n.MessageHandler(msg, peerID.String())
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

	fmt.Printf("p2p已发送消息到 %s\n", peerID)
	return nil
}

// 根据pid发送对应类型的消息
func (node *Node) SendP2PMessage(pid string, msgType uint64, data []byte) error {
	p2pMsg := &Message{
		Type: msgType,
		Data: data,
	}

	encodedMsg, err := p2pMsg.Encode()
	if err != nil {
		return fmt.Errorf("消息编码失败: %w", err)
	}

	peerID, err := peer.Decode(pid)
	if err != nil {
		return fmt.Errorf("Peer ID 解码失败: %w", err)
	}

	if err := node.SendMessage(peerID, string(encodedMsg)); err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}

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

	if peerInfo.ID == peer.ID(n.ID) {
		// 欲连接的节点为自身，return
		return nil
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
			return fmt.Errorf("p2p广播消息失败到 %s: %v", peerID, err)
		}
	}
	fmt.Println("p2p消息已广播到所有连接的节点")
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

func (node *Node) Synchronization(addr string) {
	if len(node.Peers()) < 1 {
		fmt.Println("错误: 未连接任何节点")
		return
	}
	encodedGetSnapshotmsg, err := rlp.EncodeToBytes(node.ID)
	if err != nil {
		log.Fatal("RLP encoding failed:", err)
		return
	}
	// 发送给第一个连接的节点（实际应用可扩展选择机制）
	// targetPeer := sys.p2pNode.Peers()[0]
	targetPeer, err := peer.Decode(addr)
	if err != nil {
		log.Fatal("Peer ID 解码失败: %w", err)
		return
	}
	p2pMsg := &Message{
		Type: GetSnapshotMessage,
		Data: encodedGetSnapshotmsg,
	}
	encodedMsg, err := p2pMsg.Encode()
	if err != nil {
		log.Fatal("消息编码失败: %w", err)
		return
	}
	if err := node.SendMessage(targetPeer, string(encodedMsg)); err != nil {
		log.Fatal("发送消息失败: %w", err)
		return
	}
}

type EthPeerInfo struct {
	EthAddr string
	PeerID  string
}

func (n *Node) getEthPeerInfo() *EthPeerInfo {
	// 获取当前节点的以太坊地址
	ethAddr := string(crypto.PubkeyToAddress(n.PrivateKey.PublicKey).Bytes())

	// 返回当前节点的以太坊地址和ID
	return &EthPeerInfo{
		EthAddr: ethAddr,
		PeerID:  n.ID,
	}
}

func (n *Node) HelloBroadcast() {
	// 广播Hello消息
	ethPeerInfo := n.getEthPeerInfo()
	encodedInfo, _ := rlp.EncodeToBytes(ethPeerInfo)

	p2pMsg := &Message{
		Type: HelloMessage,
		Data: encodedInfo,
	}
	encodedP2PMsg, _ := p2pMsg.Encode()
	n.BroadcastMessage(string(encodedP2PMsg))
	log.Printf("=== Hello消息已广播 ===")
}

func (n *Node) PeriodHelloBroadcast() {
	// 定时发送Hello消息

	shortInterval := 5 * time.Second
	longInterval := 2 * time.Minute

	// 第一个周期，立即执行
	select {
	case <-time.After(shortInterval):
		log.Printf("=== 第一次广播Hello消息 ===")
		n.HelloBroadcast()
	}

	log.Printf("=== Switching to long-term Hello Broadcast Mechanism (every 2 min) ===")
	// 后续周期：每隔 1 分钟运行一次
	ticker := time.NewTicker(longInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Printf("=== Hello Broadcast Mechanism (Period) ===")
			n.HelloBroadcast()
		}
	}

}
