package system

import (
	"ablockchain/cli"
	"ablockchain/consensus"
	pbftcore "ablockchain/consensus/bft/pbft/core"
	"ablockchain/consensus/pow"
	"ablockchain/core"
	"ablockchain/event"
	"ablockchain/p2p"
	"fmt"

	"log"
)

type System struct {
	p2pNode        *p2p.Node
	blockChain     *core.Blockchain
	accountManager *core.AccountManager
	consensus      consensus.Consensus
}

func StartSystem(cfg *cli.Config) *System {
	var sys System

	// 启动 P2P 节点
	node := cli.StartListen(cfg)
	sys.p2pNode = node

	// 初始化账户管理和账户
	accountManager := core.NewAccountManager()
	sys.accountManager = accountManager
	account, err := accountManager.NewAccount()
	if err != nil {
		log.Printf("cannot create new account: %v", err)
	}
	accountManager.Accounts[account.Address] = account
	// TODO: account的地址如何获取，account的各个参数如何设置，如公私钥、balance

	bc, err := core.NewBlockchain()
	if err != nil {
		log.Fatalf("initial blockchain failed: %v", err)
	}
	sys.blockChain = bc

	switch cfg.ConsensusType {
	case "pbft":
		sys.consensus = pbftcore.NewCore(node) // TODO：调用接口必须需要使用pbftcore吗？
	case "pow":
		sys.consensus = pow.NewProofOfWork(node)
	default:
		sys.consensus = pbftcore.NewCore(node)
	}

	// 开启共识模块
	sys.consensus.Start()
	log.Printf("开启共识模块……\n")

	// TODO：如对于PBFT算法，非主节点应该不需要开启挖矿进程？
	bc.StartMiner()     // 异步进程，开启判断是否要打包交易生成区块
	ListenNewBlocks(bc) // 异步进程，监听是否有新区块生成，若有则处理

	// 进入交互命令行
	commander := cli.NewCommander(node)
	commander.Run()

	return &sys
}

func ListenNewBlocks(bc *core.Blockchain) {
	consensusfinish := event.Bus.Subscribe("ConsensusFinish")

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
				}
				fmt.Println("\n##提交区块##")
				bc.AddBlock(Block)
			}
		}
	}()
}

func handleNewBlock(block *core.Block) {
	event.Bus.Publish("ConsensusStart", block)

}
