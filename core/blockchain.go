package core

import "ablockchain/storage"

type Blockchain struct {
	db     *storage.LevelDB
	TxPool *TxPool
}

func NewBlockchain() *Blockchain {
	path := "./block_storage"
	db, err := storage.NewLevelDB(path)
	if err != nil {
		return nil
	}

	txPool := NewTxPool()

	return &Blockchain{db: db, TxPool: txPool}
}

func NewGenesisBlock() *Block {
	return NewBlock(nil, nil)
}
