package main

import (
	"ablockchain/cli"
	"ablockchain/core"
)

func main() {
	// 解析命令行参数
	cfg := cli.ParseFlags()

	// 启动 P2P 节点
	node := cli.StartListen(cfg)

	// 创建区块链并启动一个异步的miner进程
	bc, err := core.NewBlockchain()
	if err != nil {
		return
	}
	bc.StartMiner()

	// 进入交互命令行
	commander := cli.NewCommander(node)
	commander.Run()
}
