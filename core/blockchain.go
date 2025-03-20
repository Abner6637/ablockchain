package core

import (
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
	db     *storage.LevelDB
	TxPool *TxPool

	currentBlockHash []byte

	NewBlockChan chan *Block
}

func NewBlockchain() (*Blockchain, error) {
	path := "./block_storage"
	db, err := storage.NewLevelDB(path)
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

	// 创建创世区块
	genesisBlock := NewGenesisBlock(genensisConfig.Difficulty)
	log.Printf("创建创世区块……\n")
	db.Put("0", genesisBlock)

	return &Blockchain{
		db:           db,
		TxPool:       txPool,
		NewBlockChan: make(chan *Block, 10), // 缓冲通道防止堵塞
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
	header := NewBlockHeader(bc.currentBlockHash, uint64(1))
	block := NewBlock(header, txs)

	// bc.AddBlock(block)
	bc.NewBlockChan <- block // 将新区块发送到通道

	bc.TxPool.ClearPackedTxs(block.Transactions)
	return block, nil
}

func (bc *Blockchain) AddBlock(block *Block) {
	str := fmt.Sprintf("%d", block.Header.Number)
	fmt.Println(str, block)
	bc.db.Put(str, block)
}
