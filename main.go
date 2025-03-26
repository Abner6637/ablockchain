package main

import (
	"ablockchain/cli"
	"ablockchain/system"
)

func main() {
	// log默认输出到stderr，将其重定向至stdout，当使用tee时可显示
	// tee默认读取stdout流
	// 或者使用tee时，前面加上”2>&1“
	//log.SetOutput(os.Stdout)

	// 解析命令行参数
	cfg := cli.ParseFlags()

	system.StartSystem(cfg)
}
