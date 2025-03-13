
main.go:
```go
package main

import (
	"ablockchain/cli"
	"ablockchain/p2p"
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
```

usage example:
```sh
# 终端1（节点A）
go run main.go --listen /ip4/0.0.0.0/tcp/9000
节点已启动 ID: QmAbc123...
监听地址: [/ip4/127.0.0.1/tcp/9000]

> connect /ip4/127.0.0.1/tcp/9001/p2p/QmXyz456...
连接成功

> send Hello NodeB!

# 终端2（节点B）
go run main.go --listen /ip4/0.0.0.0/tcp/9001
节点已启动 ID: QmXyz456...
监听地址: [/ip4/127.0.0.1/tcp/9001]

[新消息] Hello NodeB!
```