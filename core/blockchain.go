package core

import "ablockchain/storage"

type Blockchain struct {
	db     *storage.LevelDB
	TxPool *TxPool
}

func NewBlockchain() *Blockchain {
	return &Blockchain{}
}

func NewGenesisBlock() *Block {
	return NewBlock(nil, nil)
}
