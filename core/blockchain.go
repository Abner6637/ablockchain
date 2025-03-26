package core

import (
	"ablockchain/cli"
	"ablockchain/config"
	"ablockchain/storage"
	"fmt"
	"log"
	"time"
)

const BlockInterval time.Duration = 5 * time.Second
const MinTransactionsPerBlock int = 2
const MaxTransactionsPerBlock int = 10

type Blockchain struct {
	db        *storage.LevelDB
	TxPool    *TxPool
	StateDB   *StateDB // 使用Merkle Patricia Tree来存储账户状态
	StateRoot []byte   // Merkle Patricia Tree的根哈希

	CurrentBlockHash []byte
	CurBlockNum      uint64
	NewBlockChan     chan *Block
}

func NewBlockchain(cfg *cli.Config) (*Blockchain, error) {
	// path := "./block_storage"
	DBPath := cfg.DBPath
	db, err := storage.NewLevelDB(DBPath)
	if err != nil {
		return nil, err
	}

	txPool := NewTxPool()

	// 加载创世配置（需要在启动的目录下有一个创世文件
	genensisConfig, err := config.LoadGenesisConfig("./genesis.json")
	if err != nil {
		return nil, err
	}
	log.Printf("加载创世配置……\n")

	stateDB, err := NewStateDB(DBPath + "_state")
	if err != nil {
		return nil, err
	}

	// 创建创世区块
	genesisBlock := NewGenesisBlock(genensisConfig.Difficulty)
	curBlockNum := uint64(0)
	log.Printf("创建创世区块……\n")
	db.Put("0", genesisBlock)

	// 计算初始 stateRoot
	stateRoot := stateDB.trie.RootHash()

	return &Blockchain{
		db:        db,
		TxPool:    txPool,
		StateDB:   stateDB,
		StateRoot: stateRoot,
		//currentBlockHash: currentBlockHash,
		CurBlockNum:  curBlockNum,
		NewBlockChan: make(chan *Block, 10),
	}, nil
}

// 开始一个异步的miner进程
func (bc *Blockchain) StartMiner() {
	go func() {
		for {
			if bc.TxPool.PendingSize() >= MinTransactionsPerBlock {
				bc.mineNewBLock()
			}
			time.Sleep(BlockInterval)
		}
	}()
}

func (bc *Blockchain) mineNewBLock() (*Block, error) {
	txs := bc.TxPool.GetTxs()
	if len(txs) == 0 {
		return nil, fmt.Errorf("no transaction!")
	}

	// 创建新区块（该部分的difficulty需要进一步修改）
	header := NewBlockHeader(bc.CurrentBlockHash, uint64(1), bc.CurBlockNum+1)
	block := NewBlock(header, txs)

	// bc.AddBlock(block)
	bc.NewBlockChan <- block // 将新区块发送到通道

	bc.TxPool.ClearPackedTxs(block.Transactions)
	return block, nil
}

func (bc *Blockchain) AddBlock(block *Block) {
	str := fmt.Sprintf("%d", block.Header.Number)
	fmt.Println(str)
	block.PrintBlock()
	bc.db.Put(str, block)
}

func NewTestBlockchain(DBPath string) (*Blockchain, error) {
	// path := "./block_storage"
	db, err := storage.NewLevelDB(DBPath)
	if err != nil {
		return nil, err
	}

	txPool := NewTxPool()

	// 加载创世配置（需要在启动的目录下有一个创世文件
	genensisConfig, err := config.LoadGenesisConfig("./genesis.json")
	if err != nil {
		return nil, err
	}
	log.Printf("加载创世配置……\n")

	stateDB, err := NewStateDB(DBPath + "_state")
	if err != nil {
		return nil, err
	}

	// 创建创世区块
	genesisBlock := NewGenesisBlock(genensisConfig.Difficulty)
	curBlockNum := uint64(0)
	log.Printf("创建创世区块……\n")
	db.Put("0", genesisBlock)

	// 计算初始 stateRoot
	stateRoot := stateDB.trie.RootHash()

	return &Blockchain{
		db:        db,
		TxPool:    txPool,
		StateDB:   stateDB,
		StateRoot: stateRoot,
		//currentBlockHash: currentBlockHash,
		CurBlockNum:  curBlockNum,
		NewBlockChan: make(chan *Block, 10),
	}, nil
}
