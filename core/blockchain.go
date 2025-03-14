package core

import (
	"ablockchain/config"
	"ablockchain/storage"
)

type Blockchain struct {
	db     *storage.LevelDB
	TxPool *TxPool
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

func (bc *Blockchain) StartMiner() {
	go func() {

	}()
}
