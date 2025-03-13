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

	core.NewBlockchain()
	//bc := core.NewBlockchain()
	//defer bc.db.Close()

	// 进入交互命令行
	commander := cli.NewCommander(node)
	commander.Run()
}
