package p2p

import (
	"fmt"
	"testing"
	"time"
)

func TestStartNode(t *testing.T) {
	// 配置监听地址和对方节点地址
	listenAddress := "/ip4/127.0.0.1/tcp/9000"
	peerAddress := "/ip4/127.0.0.1/tcp/9001"

	// 启动节点
	node, err := StartNode(listenAddress, peerAddress)

	// 检查是否没有错误
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	// 检查返回的节点是否非空
	if node == nil {
		t.Fatal("Expected node to be non-nil")
	}
	// 检查节点的 peerAddress 是否正确
	if node.PeerAddress != peerAddress {
		t.Errorf("Expected peer address %s, got %s", peerAddress, node.PeerAddress)
	}
}

func TestConnectToPeer(t *testing.T) {
	// 配置监听地址和对方节点地址
	listenAddress1 := "/ip4/127.0.0.1/tcp/9000"
	listenAddress2 := "/ip4/127.0.0.1/tcp/9001"

	peerAddress1 := "/ip4/127.0.0.1/tcp/9001"
	peerAddress2 := "/ip4/127.0.0.1/tcp/9000"

	// 启动 Node 1
	node1, err := StartNode(listenAddress1, peerAddress1)
	if err != nil {
		t.Fatalf("Failed to start node 1: %v", err)
	}

	// 启动 Node 2
	node2, err := StartNode(listenAddress2, peerAddress2)
	if err != nil {
		t.Fatalf("Failed to start node 2: %v", err)
	}

	// 给目标节点一些时间来完全启动
	time.Sleep(2 * time.Second) // 延迟 2 秒，确保两个节点完全启动

	// 连接 Node 1 到 Node 2
	err = node1.ConnectToPeer()
	if err != nil {
		t.Errorf("Node 1 failed to connect to Node 2: %v", err)
	}

	// 连接 Node 2 到 Node 1
	err = node2.ConnectToPeer()
	if err != nil {
		t.Errorf("Node 2 failed to connect to Node 1: %v", err)
	}

	// 这里可以进一步添加验证连接是否成功的代码
	fmt.Println("Node 1 connected to Node 2 and vice versa")
}

func TestInvalidConnection(t *testing.T) {
	// 配置错误的对方节点地址
	listenAddress := "/ip4/127.0.0.1/tcp/9000"
	invalidPeerAddress := "/ip4/127.0.0.1/tcp/9999" // 不存在的地址

	// 启动节点
	node, err := StartNode(listenAddress, invalidPeerAddress)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// 尝试连接到错误的地址
	err = node.ConnectToPeer()
	if err == nil {
		t.Errorf("Expected error when connecting to invalid peer address, got nil")
	}
}
