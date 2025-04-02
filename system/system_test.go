package system

import (
	"ablockchain/core"
	"ablockchain/crypto"
	"ablockchain/p2p"
	"fmt"
	"log"
	"math/big"
	"testing"
	"time"

	"github.com/multiformats/go-multiaddr"
)

func newTestNode() (*p2p.Node, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Printf("生成密钥失败：%v", err)
		return nil, err
	}

	node, err := p2p.NewNode("/ip4/127.0.0.1/tcp/0", privateKey)
	if err != nil {
		log.Printf("启动节点1失败: %v\n", err)
		return nil, err
	}
	fmt.Printf("节点1已启动 ID: %s\n", node.ID)
	fmt.Printf("监听地址: %v\n", node.Host.Addrs())

	return node, nil
}

func TestPBFTSystem(t *testing.T) {
	node1, err := newTestNode()
	if err != nil {
		t.Fatalf("fault:%v", err)
	}

	node2, err := newTestNode()
	if err != nil {
		t.Fatalf("fault:%v", err)
	}

	// 连接节点
	node2Addr, _ := getNodeFullAddr(node2)
	if err := node1.ConnectToPeer(node2Addr); err != nil {
		t.Fatalf("连接失败: %v", err)
	}

	dbPath1 := "./test_storage1"
	bc1, err := core.NewTestBlockchain(dbPath1)

	// consensusCore1 := pbftcore.NewCore(node1)

	// consensusCore1.Start()

	// 暂时忽略交易打包成区块的过程，用TestBlock进行测试
	// bc1.StartMiner()
	ListenNewBlocks(bc1)

}

func newTestBlockForSystem() *core.Block {
	return &core.Block{
		Header: &core.BlockHeader{
			ParentHash: []byte("0df9a8f4a2f2fc354c3c8aa5e837d4db137f20ccbf3d8336e4c95ac9d0e2943e"),
			MerkleRoot: []byte("1cdfdf5680f2a639732f6aae64a8b96c10a913b46c8fcd908c9eb95925979974"),
			Time:       uint64(time.Now().Unix()),
			Difficulty: 2,
			Nonce:      0,
			Number:     big.NewInt(1),
		},
	}
}

func getNodeFullAddr(n *p2p.Node) (string, error) {
	// 获取节点的第一个 IPv4 TCP 地址
	var addr multiaddr.Multiaddr
	for _, a := range n.Host.Addrs() {
		if _, err := a.ValueForProtocol(multiaddr.P_IP4); err != nil {
			continue
		}
		if _, err := a.ValueForProtocol(multiaddr.P_TCP); err != nil {
			continue
		}
		addr = a
		break
	}
	if addr == nil {
		return "", fmt.Errorf("无可用 IPv4 TCP 地址")
	}

	// 封装为包含 PeerID 的地址
	fullAddr := addr.Encapsulate(
		multiaddr.StringCast("/p2p/" + n.Host.ID().String()),
	)
	return fullAddr.String(), nil
}
