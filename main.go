package main

import (
	"ablockchain/cli"
	"ablockchain/system"
)

func main() {
	// 解析命令行参数
	cfg := cli.ParseFlags()

	system.StartSystem(cfg)
}
