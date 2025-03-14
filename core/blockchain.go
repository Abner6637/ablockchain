package core

import (
	"ablockchain/config"
	"ablockchain/storage"
	"time"
)

const BlockInterval time.Duration = 5 * time.Second
const MinTransactionsPerBlock int = 2
const MaxTransactionsPerBlock int = 10

type Blockchain struct {
	db     *storage.LevelDB
	TxPool *TxPool

	currentBlockHash []byte
}

func NewBlockchain() (*Blockchain, error) {
	path := "./block_storage"
	db, err := storage.NewLevelDB(path)
	if err != nil {
		return nil, err
	}

	txPool := NewTxPool()

	// 加载创世配置
	genensisConfig, err := config.LoadGenesisConfig("./genesis.json")
	if err != nil {
		return nil, err
	}

	// acountManager := NewAccountManager();

	// 创建创世区块
	genesisBlock := NewGenesisBlock(genensisConfig.Difficulty)
	db.Put("0", genesisBlock)

	return &Blockchain{db: db, TxPool: txPool}, nil
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

func (bc *Blockchain) mineNewBLock() {
	txs := bc.TxPool.GetTxs()
	if len(txs) == 0 {
		return
	}

	// 创建新区块（该部分的difficulty需要进一步修改）
	header := NewBlockHeader(bc.currentBlockHash, uint64(0))
	block := NewBlock(header, txs)

	// 后续需要加入consensus部分
	//
	//
	// TODO

	bc.AddBlock(block)
	// 共识完成并将新区块加入区块链后，还需要广播该区块吗？
	//
	//
	// TODO

	bc.TxPool.ClearPackedTxs(block.Transactions)
}

func (bc *Blockchain) AddBlock(block *Block) {
}
