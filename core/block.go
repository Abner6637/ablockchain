package core

import (
	"bytes"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
)

type BlockHeader struct {
	ParentHash []byte
	Time       time.Time

	Difficulty uint64
	Number     uint64
	MerkleRoot []byte // 该区块交易的梅克尔根
	Nonce      uint32 // PoW 计算用的随机数
}

func NewBlockHeader(parentHash []byte, dif uint64) *BlockHeader {
	return &BlockHeader{
		ParentHash: parentHash,
		Time:       time.Now(),
		Difficulty: dif,
	}
}

// TODO: add other parts of bh
// Hash怎么用
func (bh *BlockHeader) BlockHash() []byte {
	var buf bytes.Buffer

	buf.Write(bh.ParentHash)

	return []byte{}
}

type Block struct {
	Header       *BlockHeader
	Transactions []*Transaction
	Hash         []byte
}

func NewBlock(bh *BlockHeader, txs []*Transaction) *Block {
	hash := bh.BlockHash()

	return &Block{
		Header:       bh,
		Transactions: txs,
		Hash:         hash,
	}
}

func (b *Block) EncodeBLock() ([]byte, error) {
	encodedBlock, err := rlp.EncodeToBytes(b)
	if err != nil {
		log.Fatal("RLP encoding failed:", err)
		return nil, err
	}
	return encodedBlock, nil
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
