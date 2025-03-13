package core

import "time"

type BlockHeader struct {
	ParentHash []byte
	Time       time.Time
}

type Block struct {
	Header       *BlockHeader
	Transactions []*Transaction
	Hash         []byte
}

func NewBlock(bh *BlockHeader, txs []*Transaction) *Block {
	return &Block{Header: bh, Transactions: txs}
}
