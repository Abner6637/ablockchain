package core

import (
	"ablockchain/config"
	"ablockchain/storage"
	"fmt"
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

// 共识部分处理mineNewBlock何时启用的逻辑？
//
// TODO
func (bc *Blockchain) mineNewBLock() (*Block, error) {
	txs := bc.TxPool.GetTxs()
	if len(txs) == 0 {
		return nil, fmt.Errorf("no transaction!")
	}

	// 创建新区块（该部分的difficulty需要进一步修改）
	header := NewBlockHeader(bc.currentBlockHash, uint64(0))
	block := NewBlock(header, txs)

	bc.AddBlock(block)

	// rlp编码区块，广播编码后的区块
	/* 这部分在共识部分书写？
	还有，很多共识并不是直接广播区块，如PBFT广播包含区块信息的Request
	encodedBlock, err := block.EncodeBLock()
	if err != nil {
		return err
	}
	p2p.BroadcastMessage(string(encodedBlock))
	*/

	// 清除已经打包到区块的交易
	bc.TxPool.ClearPackedTxs(block.Transactions)
	return block, nil
}

func (bc *Blockchain) AddBlock(block *Block) {
	str := fmt.Sprintf("%d", block.Header.Number)
	bc.db.Put(str, block) // str代表区块编号Number（可能不是这样的
}
