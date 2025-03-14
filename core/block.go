package core

import "time"

type BlockHeader struct {
	ParentHash []byte
	Time       time.Time

	Difficulty uint64
	MerkleRoot []byte // 该区块交易的梅克尔根
	Nonce      uint32 // PoW 计算用的随机数
}

type Block struct {
	Header       *BlockHeader
	Transactions []*Transaction
	Hash         []byte
}

func NewBlock(bh *BlockHeader, txs []*Transaction) *Block {
	return &Block{Header: bh, Transactions: txs}
}

func NewBlockHeader(parentHash []byte, dif uint64) *BlockHeader {
	time := time.Now()
	return &BlockHeader{
		ParentHash: parentHash,
		Time:       time,
		Difficulty: dif}
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
