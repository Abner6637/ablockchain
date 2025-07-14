package system

import (
	"ablockchain/api"
	"ablockchain/cli"
	"ablockchain/consensus"
	hotstuffcore "ablockchain/consensus/bft/hotstuff/core"
	pbftcore "ablockchain/consensus/bft/pbft/core"
	poacore "ablockchain/consensus/poa/core"

	//"ablockchain/consensus/poa"
	"ablockchain/consensus/pow"
	"ablockchain/core"
	"ablockchain/event"
	"ablockchain/p2p"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/ipfs/go-cid"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
)

type BlankValidator struct{}

func (v BlankValidator) Validate(_ string, _ []byte) error {
	return nil
}

// Add the required Select method to implement the full interface
func (v BlankValidator) Select(_ string, vals [][]byte) (int, error) {
	if len(vals) == 0 {
		return 0, fmt.Errorf("no values to select from")
	}
	return 0, nil
}

type System struct {
	p2pNode    *p2p.Node
	blockChain *core.Blockchain
	consensus  consensus.Consensus
	api        *api.API
}

func StartSystem(cfg *cli.Config) *System {
	var sys System
	// 启动 P2P 节点
	node := StartListen(cfg)
	sys.p2pNode = node

	// crypto.InitBLS() // 初始化BLS库

	if len(cfg.PeerAddrs) != 0 {
		// 阻塞两秒，等待其他节点启动再连接
		select {
		case <-time.After(2 * time.Second):
			fmt.Println("Start to connect to peers...")
		}
		for _, peerAddr := range cfg.PeerAddrs {
			sys.p2pNode.ConnectToPeer(peerAddr)
		}
	} else {
		log.Printf("peer address length: 0")
	}

	bc, err := core.NewBlockchain(cfg)
	if err != nil {
		log.Fatalf("initial blockchain failed: %v", err)
	}
	sys.blockChain = bc

	//go sys.p2pNode.PeriodHelloBroadcast() // 定时发送Hello消息

	switch cfg.ConsensusType {
	case "pbft":
		sys.consensus = pbftcore.NewCore(cfg, node)
	case "hotstuff":
		sys.consensus = hotstuffcore.NewCore(cfg, node)
	case "pow":
		sys.consensus = pow.NewProofOfWork(node)
	case "poa":
		sys.consensus = poacore.NewCore(node, sys.blockChain, cfg.ValSet)
	default:
		sys.consensus = pbftcore.NewCore(cfg, node)
	}

	log.Printf("共识类型: %s", cfg.ConsensusType)

	// 开启共识模块
	sys.consensus.Start()
	log.Printf("开启共识模块……\n")

	// 对于PBFT算法，非主节点应该不需要开启挖矿进程？
	// 但其实正常开启没有问题。只要其他节点不接受来自非主节点的区块就可以了
	sys.ListenNewBlocks()       // 异步进程，监听是否有新区块生成，若有则处理
	sys.ListenSynchronization() // 异步进程，监听是否有状态同步消息

	sys.api = api.NewAPI(bc, node)

	// 图形界面所需的RPC服务
	go StartRPCHandler(cfg.RPC, sys.api)

	if cfg.IfBackend {
		log.Printf("开启IPC Server")
		go startIPCServer(&sys)
	} else {
		// 若在前台运行，直接进入交互命令行
		cmd := NewCommander(&sys)
		cmd.Run()
	}

	return &sys
}

func StartListen(cfg *cli.Config) *p2p.Node {
	// 节点的私钥
	privateKey, libPriv := cfg.NodeKey()
	node, err := p2p.NewNode(cfg.ListenAddr, privateKey, libPriv)
	if err != nil {
		fmt.Printf("启动节点失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("节点已启动 ID: %s\n", node.ID)
	fmt.Printf("监听地址: %v\n", node.Host.Addrs())

	// 自身bootnode时
	if cfg.IsBoot == "true" {
		dhtOpts := []dht.Option{
			dht.Mode(dht.ModeServer),
			dht.ProtocolPrefix("/ablockchain"),
			dht.Validator(BlankValidator{}),
		}
		kad, err := dht.New(context.Background(), node.Host, dhtOpts...)
		if err != nil {
			log.Printf("Failed to create DHT: %v", err)
			return node
		}

		if err := kad.Bootstrap(context.Background()); err != nil {
			log.Printf("Failed to bootstrap DHT: %v", err)
		}

		RegisterMyselfInDHT(kad)
		log.Printf("启动Bootnode DHT")
	}

	// 自身非bootnode，且需要连接bootnode时，说明开启了DHT连接
	if cfg.BootNode != "" && cfg.IsBoot == "false" {
		ctx := context.Background()

		// 先连接bootnode
		bootAddr, err := multiaddr.NewMultiaddr(cfg.BootNode)
		if err != nil {
			log.Printf("Failed to parse bootnode address: %v", err)
			return node
		}

		bootInfo, err := peer.AddrInfoFromP2pAddr(bootAddr)
		if err != nil {
			log.Printf("Failed to extract bootnode info: %v", err)
			return node
		}

		log.Printf("开始连接bootnode %s...", bootInfo.ID)
		if err := node.Host.Connect(ctx, *bootInfo); err != nil {
			log.Printf("Failed to connect to bootnode: %v", err)
			return node
		}

		// 初始化 DHT
		dhtOpts := []dht.Option{
			dht.Mode(dht.ModeAuto),
			dht.BootstrapPeers(*bootInfo),
			dht.ProtocolPrefix("/ablockchain"),
			dht.Validator(BlankValidator{}),
		}

		kad, err := dht.New(ctx, node.Host, dhtOpts...)
		if err != nil {
			log.Printf("Failed to create DHT: %v", err)
			return node
		}

		log.Printf("Bootstrapping DHT...")
		if err := kad.Bootstrap(ctx); err != nil {
			log.Printf("Failed to bootstrap DHT: %v", err)
		}

		// 开始 discovery
		/* 		DiscoverAndConnectToOthers(kad, node.Host)
		   		RegisterMyselfInDHT(kad)
		   		DiscoverAndConnectToOthers(kad, node.Host) */
		go StartAutoDiscovery(ctx, kad, node.Host)
	}
	return node
}

func (sys *System) ListenNewBlocks() {
	consensusfinish := event.Bus.Subscribe("ConsensusFinish")
	// publishTx := event.Bus.Subscribe("PublishTx")
	go func() {
		for {
			select {
			// 打包区块（触发共识）
			case block := <-sys.blockChain.NewBlockChan:
				fmt.Println("\n##触发共识##")
				handleNewBlock(block)
			// 提交区块（上链）
			case msg := <-consensusfinish:
				Block, ok := msg.(*core.Block)
				if !ok {
					log.Fatal("转换失败: 事件数据不是 *core.Block 类型")
					continue
				}
				fmt.Println("\n##提交区块##")
				if !sys.blockChain.AddBlock(Block) {
					fmt.Println("\n区块信息与当前状态不匹配，触发同步！")
					sys.RandomSyc()
				}

				// case signtx := <-publishTx:
				// 	txBytes, ok := signtx.([]byte)
				// 	if !ok {
				// 		log.Fatal("收到无效的交易数据，期望类型: []byte")
				// 		continue
				// 	}
				// 	fmt.Println("\n##广播新交易##")
				// 	handlePublishTx(txBytes)
			}
		}
	}()
}

func handleNewBlock(block *core.Block) {
	event.Bus.Publish("ConsensusStart", block)
}

func startIPCServer(sys *System) {
	socketPath := "/tmp/ablockchain.sock"
	os.Remove(socketPath) // 防止已有的阻塞

	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("监听 socket 失败: %v", err)
	}
	defer ln.Close()

	fmt.Println("交互接口监听中：", socketPath)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		cmd := NewIPCCommander(sys, conn, conn)
		go cmd.Run()
	}
}

func RegisterMyselfInDHT(kad *dht.IpfsDHT) {
	content := []byte("ablockchain-network")
	hash, _ := multihash.Sum(content, multihash.SHA2_256, -1)
	c := cid.NewCidV1(cid.Raw, hash)

	ctx := context.Background()
	if err := kad.Provide(ctx, c, true); err != nil {
		log.Printf("Failed to register in DHT: %v", err)
		return
	}
	log.Printf("成功注册 DHT")
}

func DiscoverAndConnectToOthers(kad *dht.IpfsDHT, h host.Host) {
	ctx := context.Background()
	content := []byte("ablockchain-network")
	hash, _ := multihash.Sum(content, multihash.SHA2_256, -1)
	c := cid.NewCidV1(cid.Raw, hash)

	log.Printf("Starting peer discovery process...")

	// 通过provider查找连接peer
	peerChan := kad.FindProvidersAsync(ctx, c, 20)
	for peer := range peerChan {
		log.Printf("Found provider peer: %s", peer.ID)
		if peer.ID == h.ID() {
			continue
		}
		tryConnect(ctx, h, peer)
	}

	/* // Second try: Query all peers in routing table
	log.Printf("Querying routing table...")
	for _, p := range kad.RoutingTable().ListPeers() {
		log.Printf("Found routing table peer: %s", p)
		if p == h.ID() {
			continue
		}

		// Get peer info from peerstore
		addrs := h.Peerstore().Addrs(p)
		if len(addrs) == 0 {
			log.Printf("No addresses found for peer %s", p)
			continue
		}

		peerInfo := peer.AddrInfo{
			ID:    p,
			Addrs: addrs,
		}
		tryConnect(ctx, h, peerInfo)
	}

	// Third try: Find closest peers
	log.Printf("Finding closest peers...")
	if peers, err := kad.GetClosestPeers(ctx, "/ablockchain"); err == nil {
		for _, p := range peers {
			log.Printf("Found closest peer: %s", p)
			if p == h.ID() {
				continue
			}
			peerInfo := h.Peerstore().PeerInfo(p)
			tryConnect(ctx, h, peerInfo)
		}
	} */
}

func tryConnect(ctx context.Context, h host.Host, pi peer.AddrInfo) {
	if h.Network().Connectedness(pi.ID) == network.Connected {
		log.Printf("Already connected to peer: %s", pi.ID)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log.Printf("Attempting to connect to peer %s with addresses: %v", pi.ID, pi.Addrs)
	if err := h.Connect(ctx, pi); err != nil {
		log.Printf("Failed to connect to peer %s: %v", pi.ID, err)
		return
	}
	log.Printf("Successfully connected to peer: %s", pi.ID)
}

func StartAutoDiscovery(ctx context.Context, kad *dht.IpfsDHT, h host.Host) {
	shortInterval := 2 * time.Second
	longInterval := 1 * time.Minute
	firstInterval := 0 * time.Second

	// 第一个周期，立即执行
	select {
	case <-ctx.Done():
		return
	case <-time.After(firstInterval):
		log.Printf("=== [Quick Startup] Discovery round 1 ===")
		RegisterMyselfInDHT(kad)
		DiscoverAndConnectToOthers(kad, h)
	}
	// 第二、三个周期，隔2秒运行
	for i := 0; i < 2; i++ {
		select {
		case <-ctx.Done():
			return
		case <-time.After(shortInterval):
			log.Printf("=== [Quick Startup] Discovery round %d ===", i+1)
			RegisterMyselfInDHT(kad)
			DiscoverAndConnectToOthers(kad, h)

			// 打印连接节点
			peers := h.Network().Peers()
			log.Printf("=== Connected peers: %d ===", len(peers))
			for _, p := range peers {
				conns := h.Network().ConnsToPeer(p)
				for _, conn := range conns {
					log.Printf("  - %s via %s", p, conn.RemoteMultiaddr())
				}
			}
		}
	}

	log.Printf("=== Switching to long-term discovery (every 1 min) ===")
	// 后续周期：每隔 1 分钟运行一次
	ticker := time.NewTicker(longInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Printf("=== Periodic Discovery ===")
			RegisterMyselfInDHT(kad)
			DiscoverAndConnectToOthers(kad, h)

			// 打印连接节点
			peers := h.Network().Peers()
			log.Printf("=== Connected peers: %d ===", len(peers))
			for _, p := range peers {
				conns := h.Network().ConnsToPeer(p)
				for _, conn := range conns {
					log.Printf("  - %s via %s", p, conn.RemoteMultiaddr())
				}
			}
		}
	}
}

func StartRPCHandler(port string, a *api.API) {
	http.HandleFunc("/", a.Handler)

	log.Println("RPC server listening on :" + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
