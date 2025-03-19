package core

import (
	"ablockchain/crypto"
	"bytes"
	"encoding/binary"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
)

type BlockHeader struct {
	ParentHash []byte
	Time       time.Time

	Difficulty uint64
	Number     uint64
	MerkleRoot []byte
	Nonce      uint64 // PoW 计算用的随机数
}

func NewBlockHeader(parentHash []byte, dif uint64) *BlockHeader {
	return &BlockHeader{
		ParentHash: parentHash,
		Time:       time.Now(),
		Difficulty: dif,
	}
}

// 计算区块头哈希
func (bh *BlockHeader) BlockHash() []byte {
	var buf bytes.Buffer

	buf.Write(bh.ParentHash)
	binary.Write(&buf, binary.BigEndian, uint32(bh.Time.Unix()))
	binary.Write(&buf, binary.BigEndian, bh.Difficulty)
	binary.Write(&buf, binary.BigEndian, bh.Number)
	buf.Write(bh.MerkleRoot)
	binary.Write(&buf, binary.BigEndian, bh.Nonce)

	hash := crypto.GlobalHashAlgorithm.Hash(buf.Bytes())
	return hash
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

func DecodeBlock(data []byte) (*Block, error) {
	var bk Block
	err := rlp.DecodeBytes(data, &bk)
	if err != nil {
		log.Fatal("Failed to decode RLP data:", err)
		return nil, err
	}
	return &bk, nil
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
