package cli

import (
	"ablockchain/crypto"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"log"
	"os"
)

type Config struct {
	ListenAddr    string //p2p监听地址
	ConsensusType string //共识类型
	DBPath        string //数据库路径
	ConsensusNum  uint64

	NodeKeyFile string
}

func ParseFlags() *Config {
	cfg := &Config{}
	// 添加 -db 选项，默认为 "./block_storage"
	flag.StringVar(&cfg.DBPath, "db", "./block_storage", "数据库存储路径")
	flag.StringVar(&cfg.ListenAddr, "listen", "/ip4/0.0.0.0/tcp/0", "监听地址")
	flag.StringVar(&cfg.ConsensusType, "consensus", "pbft", "共识协议:pow||pbft")
	flag.StringVar(&cfg.NodeKeyFile, "nodekeyfilepath", "nodekey", "共识协议:pow||pbft")
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	return cfg
}

func (c *Config) NodeKey() *ecdsa.PrivateKey {
	keyfile := c.NodeKeyFile
	if key, err := crypto.LoadECDSA(keyfile); err == nil {
		return key
	}

	key, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf(fmt.Sprintf("Failed to generate node key: %v", err))
	}

	keyfile = "nodekey"
	if err := crypto.SaveECDSA(keyfile, key); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to persist node key: %v", err))
	}

	return key
}
