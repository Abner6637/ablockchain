package system

import (
	"ablockchain/cli"
	"ablockchain/core"
	"ablockchain/p2p"

	"log"
)

type System struct {
	p2pNode    *p2p.Node
	blockChain *core.Blockchain
}

func StartSystem(cfg *cli.Config) *System {
	// 启动 P2P 节点
	node := cli.StartListen(cfg)

	bc, err := core.NewBlockchain()
	if err != nil {
		log.Fatalf("initial blockchain failed")
	}

	bc.StartMiner()     // 开启判断是否要打包交易生成区块
	ListenNewBlocks(bc) // 监听是否有新区块生成

	// 进入交互命令行
	commander := cli.NewCommander(node)
	commander.Run()

	return &System{
		p2pNode:    node,
		blockChain: bc,
	}
}

// 监听通道，是否有新区块产生
// 后续可以增加其他监听通道内容在里面
//
// TODO
func ListenNewBlocks(bc *core.Blockchain) {
	go func() {
		for {
			select {
			case block := <-bc.NewBlockChan:
				handleNewBlock(block)
			}
		}
	}()
}

// 处理新区块
// 1. 广播区块（可能不需要，在共识里面操作？）
// 2. 触发一个事件，发送一个消息给共识模块，告知有新区块生成；相应地，共识部分需要增加监听该事件的内容
//
// TODO
func handleNewBlock(block *core.Block) {

}
