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
	ListenAddr    string   // p2p监听地址
	ConsensusType string   // 共识类型
	DBPath        string   // 数据库路径
	ConsensusNum  int      // 参加共识节点的数目
	NodeKeyFile   string   // 存储节点私钥的文件
	AddressFile   string   // 存储节点地址的文件
	P2PIDFile     string   // 存储节点P2P ID的文件
	ValSet        []string // 参加共识节点的地址集合
	IfBackend     bool     // 是否以后台模式启动
}

// 定义一个自定义类型，用于支持多次传入的字符串参数
type stringSlice []string

// String 实现了 flag.Value 接口，返回当前存储的字符串切片的表示
func (s *stringSlice) String() string {
	return fmt.Sprint(*s)
}

// Set 实现了 flag.Value 接口，将传入的参数追加到切片中
func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func ParseFlags() (*Config, bool) {
	cfg := &Config{}
	// 添加 -db 选项，默认为 "./block_storage"
	flag.StringVar(&cfg.DBPath, "db", "./block_storage", "数据库存储路径")
	flag.StringVar(&cfg.ListenAddr, "listen", "/ip4/0.0.0.0/tcp/0", "监听地址")
	flag.StringVar(&cfg.ConsensusType, "consensus", "pbft", "共识协议: pow || pbft")
	flag.IntVar(&cfg.ConsensusNum, "consensusnum", 4, "共识节点数目")
	flag.StringVar(&cfg.NodeKeyFile, "nodekey", "./key_store/nodekey", "私钥存储地址")
	flag.StringVar(&cfg.AddressFile, "address", "./key_store/address", "节点地址存储地址")
	flag.StringVar(&cfg.P2PIDFile, "p2p", "./key_store/p2pID", "节点p2p ID存储地址")
	flag.BoolVar(&cfg.IfBackend, "backend", false, "是否以后台方式运行")

	// 定义一个自定义类型变量，用于接收额外的字符串参数
	var valAddresses stringSlice
	flag.Var(&valAddresses, "valaddress", "参与共识的节点地址，以'0x'开头，可以多次指定，比如 -valaddress param1 -valaddress param2")

	flag.Usage = func() {
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
	}

	attach := flag.Bool("attach", false, "附加到已运行的节点进程")

	flag.Parse()

	cfg.ValSet = valAddresses

	for _, arg := range os.Args[1:] {
		if arg == "attach" {
			*attach = true
		}
	}

	return cfg, *attach
}

func (c *Config) NodeKey() *ecdsa.PrivateKey {
	keyfile := c.NodeKeyFile
	log.Printf("keyfile path: %s", keyfile)
	dir := filepath.Dir(keyfile)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil { // 创建目录
		log.Fatalf("Failed to create key directory %s: %v", dir, err)
	}

	if key, err := crypto.LoadECDSA(keyfile); err == nil {
		c.WriteAddress(key)

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

	c.WriteAddress(key)

	return key
}

func (c *Config) WriteAddress(key *ecdsa.PrivateKey) {
	addressfile := c.AddressFile
	addDir := filepath.Dir(addressfile)
	if err := os.MkdirAll(addDir, os.ModePerm); err != nil { // 创建目录
		log.Fatalf("Failed to create key directory %s: %v", addDir, err)
	}

	address := crypto.PubkeyToAddress(key.PublicKey).Bytes()
	hexAddress := fmt.Sprintf("0x%x", address)
	if err := os.WriteFile(addressfile, []byte(hexAddress), 0644); err != nil {
		log.Fatalf("Failed to persist node address: %v\n", err)
	}
}
