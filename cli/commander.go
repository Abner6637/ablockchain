package cli

import (
	"ablockchain/p2p"
	"flag"
	"fmt"
	"os"
)

type Config struct {
	ListenAddr string
}

func ParseFlags() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.ListenAddr, "listen", "/ip4/0.0.0.0/tcp/0", "监听地址")
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	return cfg
}

func StartListen(cfg *Config) *p2p.Node {
	node, err := p2p.StartListen(cfg.ListenAddr)
	if err != nil {
		fmt.Printf("启动节点失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("节点已启动 ID: %s\n", node.ID)
	fmt.Printf("监听地址: %v\n", node.Host.Addrs())
	return node
}
