package system

import (
	"ablockchain/cli"
	"ablockchain/core"
	"ablockchain/p2p"
	"fmt"

	"log"
)

type System struct {
	p2pNode        *p2p.Node
	blockChain     *core.Blockchain
	accountManager *core.AccountManager
}

func StartSystem(cfg *cli.Config) *System {
	// 启动 P2P 节点
	node := cli.StartListen(cfg)

	// 初始化账户管理和账户
	accountManager := core.NewAccountManager()
	account, err := accountManager.NewAccount()
	if err != nil {
		fmt.Errorf("cannot create new account")
	}
	accountManager.Accounts[account.Address] = account
	// TODO: account的地址如何获取，account的各个参数如何设置，如公私钥、balance

	bc, err := core.NewBlockchain()
	if err != nil {
		log.Fatalf("initial blockchain failed")
	}

	bc.StartMiner()     // 异步进程，开启判断是否要打包交易生成区块
	ListenNewBlocks(bc) // 异步进程，监听是否有新区块生成，若有则处理

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
