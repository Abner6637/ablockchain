package main

import (
	"ablockchain/cli"
)

func main() {
	// 解析命令行参数
	cfg := cli.ParseFlags()

	// 启动 P2P 节点
	node := cli.StartListen(cfg)

	// 进入交互命令行
	commander := cli.NewCommander(node)
	commander.Run()
}
