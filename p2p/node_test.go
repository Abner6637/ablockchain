package p2p

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
)

// 测试节点启动功能
func TestStartListen(t *testing.T) {
	// 使用随机端口（tcp/0）
	listenAddr := "/ip4/127.0.0.1/tcp/0"
	node, err := NewNode(listenAddr)
	if err != nil {
		t.Fatalf("启动节点失败: %v", err)
	}
	defer node.Host.Close()

	// 验证基础属性
	if node.Host == nil {
		t.Error("Host 不应为 nil")
	}
	if node.ID == "" {
		t.Error("PeerID 不应为空")
	}

	// 验证至少有一个有效的 IPv4 TCP 地址
	foundValidAddr := false
	for _, addr := range node.Host.Addrs() {
		// 检查 IPv4
		if _, err := addr.ValueForProtocol(multiaddr.P_IP4); err != nil {
			continue
		}
		// 检查 TCP
		if _, err := addr.ValueForProtocol(multiaddr.P_TCP); err != nil {
			continue
		}
		foundValidAddr = true
		break
	}
	if !foundValidAddr {
		t.Error("未找到有效的 IPv4 TCP 地址")
	}
}

// 测试正常节点连接
func TestPeerConnection(t *testing.T) {
	// 启动两个节点
	nodeA, err := NewNode("/ip4/127.0.0.1/tcp/0")
	if err != nil {
		t.Fatalf("启动 nodeA 失败: %v", err)
	}
	defer nodeA.Host.Close()

	nodeB, err := NewNode("/ip4/127.0.0.1/tcp/0")
	if err != nil {
		t.Fatalf("启动 nodeB 失败: %v", err)
	}
	defer nodeB.Host.Close()

	// 获取 nodeB 的完整地址（包含 PeerID）
	nodeBAddr, err := getNodeFullAddr(nodeB)
	if err != nil {
		t.Fatalf("获取 nodeB 地址失败: %v", err)
	}

	// 连接 nodeA 到 nodeB
	if err := nodeA.ConnectToPeer(nodeBAddr); err != nil {
		t.Fatalf("连接失败: %v", err)
	}

	// 验证连接状态
	connStatus := nodeA.Host.Network().Connectedness(nodeB.Host.ID())
	if connStatus != network.Connected {
		t.Errorf("连接状态异常，期望 Connected，实际为 %v", connStatus)
	}
}

// 测试连接无效地址
func TestConnectToInvalidAddress(t *testing.T) {
	node, err := NewNode("/ip4/127.0.0.1/tcp/0")
	if err != nil {
		t.Fatalf("启动节点失败: %v", err)
	}
	defer node.Host.Close()

	// 使用无效的 multiaddr 格式
	testCases := []struct {
		name    string
		address string
	}{
		{"无效格式", "invalid_multiaddr"},
		{"正确格式但不存在节点", "/ip4/127.0.0.1/tcp/12345/p2p/QmInvalidPeerID"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := node.ConnectToPeer(tc.address)
			if err == nil {
				t.Error("预期返回错误，但结果为 nil")
			}
		})
	}
}

func TestMessageExchange(t *testing.T) {
	// 启动两个节点
	nodeA, _ := NewNode("/ip4/127.0.0.1/tcp/0")
	defer nodeA.Host.Close()

	nodeB, _ := NewNode("/ip4/127.0.0.1/tcp/0")
	defer nodeB.Host.Close()

	// 设置消息接收验证
	received := make(chan string)
	nodeB.SetMessageHandler(func(msg string) {
		received <- msg
	})

	// 连接节点
	nodeBAddr, _ := getNodeFullAddr(nodeB)
	if err := nodeA.ConnectToPeer(nodeBAddr); err != nil {
		t.Fatalf("连接失败: %v", err)
	}

	// 发送测试消息
	testMsg := "Hello libp2p!"
	if err := nodeA.SendMessage(nodeB.Host.ID(), testMsg); err != nil {
		t.Fatalf("发送失败: %v", err)
	}

	// 验证接收
	select {
	case msg := <-received:
		if msg != testMsg {
			t.Errorf("消息不匹配，期望: %s，实际: %s", testMsg, msg)
		}
	case <-time.After(5 * time.Second):
		t.Error("未收到消息")
	}
}

func TestBroadcast(t *testing.T) {
	// 启动3个节点
	nodeA, _ := NewNode("/ip4/127.0.0.1/tcp/0")
	defer nodeA.Host.Close()

	nodeB, _ := NewNode("/ip4/127.0.0.1/tcp/0")
	defer nodeB.Host.Close()

	nodeC, _ := NewNode("/ip4/127.0.0.1/tcp/0")
	defer nodeC.Host.Close()

	// 设置消息接收验证
	receivedB := make(chan string)
	nodeB.SetMessageHandler(func(msgB string) {
		receivedB <- msgB
	})

	receivedC := make(chan string)
	nodeC.SetMessageHandler(func(msgC string) {
		receivedC <- msgC
	})

	// 连接节点
	nodeBAddr, _ := getNodeFullAddr(nodeB)
	if err := nodeA.ConnectToPeer(nodeBAddr); err != nil {
		t.Fatalf("连接失败: %v", err)
	}
	nodeCAddr, _ := getNodeFullAddr(nodeC)
	if err := nodeA.ConnectToPeer(nodeCAddr); err != nil {
		t.Fatalf("连接失败: %v", err)
	}

	// 发送测试消息
	testMsg := "Hello libp2p!"
	if err := nodeA.BroadcastMessage(testMsg); err != nil {
		t.Fatalf("发送失败: %v", err)
	}

	// 使用WaitGroup等待两个节点的消息验证
	var wg sync.WaitGroup
	wg.Add(2) // 等待两个节点

	// 在并发协程中验证消息
	go func() {
		defer wg.Done()
		msgB := <-receivedB
		if msgB != testMsg {
			t.Errorf("nodeB: 消息不匹配，期望: %s，实际: %s", testMsg, msgB)
		}
	}()

	go func() {
		defer wg.Done()
		msgC := <-receivedC
		if msgC != testMsg {
			t.Errorf("nodeC: 消息不匹配，期望: %s，实际: %s", testMsg, msgC)
		}
	}()

	// 等待两个消息接收的协程完成
	wg.Wait()
}

// 辅助函数：获取节点的完整 multiaddr 地址（含 PeerID）
func getNodeFullAddr(n *Node) (string, error) {
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
