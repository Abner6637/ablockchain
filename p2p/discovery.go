package p2p

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// 自定义 Notifee 结构实现发现回调
type discoveryNotifee struct {
	host host.Host
}

// 当发现新节点时触发此方法
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("Discovered peer: %s\n", pi.ID)

	// 避免连接自己
	if pi.ID == n.host.ID() {
		return
	}

	// 尝试连接目标节点
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if n.host.Network().Connectedness(pi.ID) != network.Connected {
		if err := n.host.Connect(ctx, pi); err != nil {
			fmt.Printf("Failed to connect to %v: %v\n", pi.Addrs, err)
		} else {
			fmt.Printf("Successfully connected to %v\n", pi.Addrs)
		}
	}
}

// 设置 mDNS 服务
func setupMDNS(h host.Host, serviceTag string) {
	service := mdns.NewMdnsService(h, serviceTag, &discoveryNotifee{host: h})

	// 启动服务
	if err := service.Start(); err != nil {
		panic(err)
	}
	defer service.Close()
	select {
	case <-time.After(2 * time.Second):
		fmt.Println("Stopping mDNS discovery after 2 seconds...")
		return
	}
}
