package core

import "time"

type BlockHeader struct {
	ParentHash []byte
	Time       time.Time
	Difficulty uint64
}

type Block struct {
	Header       *BlockHeader
	Transactions []*Transaction
	Hash         []byte
}

func NewBlock(bh *BlockHeader, txs []*Transaction) *Block {
	return &Block{Header: bh, Transactions: txs}
}

func NewGenesisBlock(dif uint64) *Block {
	return &Block{
		Header: &BlockHeader{
			ParentHash: []byte("0"),
			Time:       time.Now(),
			Difficulty: dif,
		},
	}
}
