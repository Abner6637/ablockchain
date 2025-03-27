package cli

import (
	"ablockchain/crypto"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	ListenAddr    string //p2p监听地址
	ConsensusType string //共识类型
	DBPath        string //数据库路径
	ConsensusNum  int
	NodeKeyFile   string
}

func ParseFlags() *Config {
	cfg := &Config{}
	// 添加 -db 选项，默认为 "./block_storage"
	flag.StringVar(&cfg.DBPath, "db", "./block_storage", "数据库存储路径")
	flag.StringVar(&cfg.ListenAddr, "listen", "/ip4/0.0.0.0/tcp/0", "监听地址")
	flag.StringVar(&cfg.ConsensusType, "consensus", "pbft", "共识协议: pow || pbft")
	flag.IntVar(&cfg.ConsensusNum, "consensusnum", 4, "共识节点数目")
	flag.StringVar(&cfg.NodeKeyFile, "nodekey", "./key_store/nodekey", "私钥存储地址")
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	return cfg
}

func (c *Config) NodeKey() *ecdsa.PrivateKey {
	keyfile := c.NodeKeyFile

	dir := filepath.Dir(keyfile)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil { // 创建目录
		log.Fatalf("Failed to create key directory %s: %v", dir, err)
	}

	if key, err := crypto.LoadECDSA(keyfile); err == nil {
		return key
	}

	key, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf(fmt.Sprintf("Failed to generate node key: %v", err))
	}

	keyfile = c.NodeKeyFile
	if err := crypto.SaveECDSA(keyfile, key); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to persist node key: %v", err))
	}

	return key
}
