package system

import (
	"ablockchain/cli"
	"ablockchain/consensus"
	pbftcore "ablockchain/consensus/bft/pbft/core"
	"ablockchain/consensus/pow"
	"ablockchain/core"
	"ablockchain/event"
	"ablockchain/p2p"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/ipfs/go-cid"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multihash"
)

type System struct {
	p2pNode    *p2p.Node
	blockChain *core.Blockchain
	// accountManager *core.AccountManager
	consensus consensus.Consensus
}

func StartSystem(cfg *cli.Config) *System {
	var sys System

	// 启动 P2P 节点
	node := StartListen(cfg)
	sys.p2pNode = node

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

	// // 初始化账户管理和账户
	// accountManager := core.NewAccountManager()
	// sys.accountManager = accountManager
	// account, err := accountManager.NewAccount()
	// if err != nil {
	// 	log.Printf("cannot create new account: %v", err)
	// }
	// accountManager.Accounts[account.Address] = account

	bc, err := core.NewBlockchain(cfg)
	if err != nil {
		log.Fatalf("initial blockchain failed: %v", err)
	}
	sys.blockChain = bc

	switch cfg.ConsensusType {
	case "pbft":
		sys.consensus = pbftcore.NewCore(cfg, node)
	case "pow":
		sys.consensus = pow.NewProofOfWork(node)
	default:
		sys.consensus = pbftcore.NewCore(cfg, node)
	}

	// 开启共识模块
	sys.consensus.Start()
	log.Printf("开启共识模块……\n")

	// 对于PBFT算法，非主节点应该不需要开启挖矿进程？
	// 但其实正常开启没有问题。只要其他节点不接受来自非主节点的区块就可以了
	// bc.StartMiner("")     // 异步进程，开启判断是否要打包交易生成区块
	ListenNewBlocks(bc) // 异步进程，监听是否有新区块生成，若有则处理

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

	/*
		if cfg.IsBoot == "true" {
			kad, _ := dht.New(context.Background(), node.Host, dht.Mode(dht.ModeServer))
			_ = kad.Bootstrap(context.Background())
			RegisterMyselfInDHT(kad)
			log.Printf("bootnode启动DHT")
		}

		if cfg.BootNode != "" && cfg.IsBoot == "false" {
			ctx := context.Background()
			log.Printf("bootnode addr:%s", cfg.BootNode)
			bootAddr, err := multiaddr.NewMultiaddr(cfg.BootNode)
			if err != nil {
				log.Printf("failed to parse peer address: %v", err)
			}
			bootInfo, err := peer.AddrInfoFromP2pAddr(bootAddr)
			if err != nil {
				log.Printf("failed to extract peer info: %v", err)
			}
			if err := node.Host.Connect(ctx, *bootInfo); err != nil {
				log.Printf("无法连接bootnode：%v", err)
			}
			log.Printf("连接bootnode")
			kad, _ := dht.New(ctx, node.Host)
			_ = kad.Bootstrap(ctx)
			RegisterMyselfInDHT(kad)
			log.Printf("启动DHT")

			log.Printf("start a discovery")
			go DiscoverAndConnectToOthers(kad, node.Host)
			go StartAutoDiscovery(ctx, kad, node.Host, time.Duration(15*time.Second))
		}
	*/

	return node
}

func ListenNewBlocks(bc *core.Blockchain) {
	consensusfinish := event.Bus.Subscribe("ConsensusFinish")
	// publishTx := event.Bus.Subscribe("PublishTx")
	go func() {
		for {
			select {
			// 打包区块（触发共识）
			case block := <-bc.NewBlockChan:
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
				bc.AddBlock(Block)
				bc.CurBlockNum = Block.Header.Number

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
	content := []byte("libp2p-group")
	hash, _ := multihash.Sum(content, multihash.SHA2_256, -1)
	cid := cid.NewCidV1(cid.Raw, hash)
	kad.Provide(context.Background(), cid, true)
}

func DiscoverAndConnectToOthers(kad *dht.IpfsDHT, h host.Host) {
	content := []byte("libp2p-group")
	hash, _ := multihash.Sum(content, multihash.SHA2_256, -1)
	cid := cid.NewCidV1(cid.Raw, hash)
	peers := kad.FindProvidersAsync(context.Background(), cid, 50)
	for p := range peers {
		if p.ID == h.ID() {
			continue
		}
		h.Peerstore().AddAddrs(p.ID, p.Addrs, peerstore.TempAddrTTL)
		if h.Network().Connectedness(p.ID) != network.Connected {
			fmt.Println("Connecting to discovered peer:", p.ID)
			if err := h.Connect(context.Background(), p); err != nil {
				fmt.Println("Connect failed:", err)
			}
		}
	}
}

func StartAutoDiscovery(ctx context.Context, kad *dht.IpfsDHT, h host.Host, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("🔌 AutoDiscovery stopped.")
			return
		case <-ticker.C:
			log.Printf("start a discovery")
			DiscoverAndConnectToOthers(kad, h)
		}
	}
}
