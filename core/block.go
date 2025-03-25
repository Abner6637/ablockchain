package core

import (
	"ablockchain/crypto"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
)

// TODO:修改time类型支持rlp编码
type BlockHeader struct {
	ParentHash []byte
	Time       uint64

	Difficulty uint64
	Number     uint64
	MerkleRoot []byte
	Nonce      uint64 // PoW 计算用的随机数
}

func NewBlockHeader(parentHash []byte, dif uint64, num uint64) *BlockHeader {
	return &BlockHeader{
		ParentHash: parentHash,
		Time:       uint64(time.Now().Unix()),
		Difficulty: dif,
		Number:     num,
	}
}

// 计算区块头哈希
func (bh *BlockHeader) BlockHash() []byte {
	var buf bytes.Buffer

	buf.Write(bh.ParentHash)
	binary.Write(&buf, binary.BigEndian, bh.Time)
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

func (b *Block) EncodeBlock() ([]byte, error) {
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
			Time:       uint64(time.Now().Unix()),
			Difficulty: dif,
			Number:     0,
		},
	}
}

func (b *Block) PrintBlock() {
	if b.Header == nil {
		fmt.Println("BlockHeader is nil")
		return
	}

	fmt.Printf("BlockHash: %x\n", b.Hash)

	fmt.Println("BlockHeader:")
	fmt.Printf("  ParentHash: %x\n", b.Header.ParentHash) // 输出字节数组为十六进制
	fmt.Printf("  Time: %v\n", b.Header.Time)
	fmt.Printf("  Difficulty: %d\n", b.Header.Difficulty)
	fmt.Printf("  Number: %d\n", b.Header.Number)
	fmt.Printf("  MerkleRoot: %x\n", b.Header.MerkleRoot)
	fmt.Printf("  Nonce: %d\n", b.Header.Nonce)
}
