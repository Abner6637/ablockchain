package core

import (
	"ablockchain/crypto"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
)

type BlockHeader struct {
	ParentHash []byte
	Time       uint64

	Difficulty uint64
	Number     *big.Int
	MerkleRoot []byte
	Nonce      uint64 // PoW 计算用的随机数
	Coinbase   string // 矿工地址，用于接收奖励

	Extra []byte
}

func NewBlockHeader(parentHash []byte, dif uint64, num *big.Int, merkleroot []byte, coinbase string) *BlockHeader {
	return &BlockHeader{
		ParentHash: parentHash,
		Time:       uint64(time.Now().Unix()),
		Difficulty: dif,
		Number:     num,
		MerkleRoot: merkleroot,
		Coinbase:   coinbase,
	}
}

func (h *BlockHeader) MarshalJSON() ([]byte, error) {
	type Alias BlockHeader
	return json.Marshal(&struct {
		*Alias
		ParentHash string `json:"ParentHash"`
		MerkleRoot string `json:"MerkleRoot"`
		Number     uint64 `json:"Number"` // big.Int 需要转成 uint64
	}{
		Alias:      (*Alias)(h),
		ParentHash: fmt.Sprintf("%x", h.ParentHash),
		MerkleRoot: fmt.Sprintf("%x", h.MerkleRoot),
		Number:     h.Number.Uint64(),
	})
}

// 计算区块头哈希
func (bh *BlockHeader) BlockHash() []byte {
	data := bytes.Join(
		[][]byte{
			bh.ParentHash,
			bh.MerkleRoot,
			[]byte(fmt.Sprintf("%v", bh.Time)),
			[]byte(fmt.Sprintf("%d", bh.Difficulty)),
			[]byte(fmt.Sprintf("%d", bh.Number)),
			[]byte(fmt.Sprintf("%d", bh.Nonce)),
		},
		[]byte{},
	)
	hash := crypto.GlobalHashAlgorithm.Hash(data)
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

func (b *Block) MarshalJSON() ([]byte, error) {
	type Alias Block // 创建一个别名类型，避免递归调用 MarshalJSON
	return json.Marshal(&struct {
		*Alias
		Hash string `json:"Hash"`
	}{
		Alias: (*Alias)(b),
		Hash:  fmt.Sprintf("%x", b.Hash), // 将哈希转换为十六进制字符串
	})
}

func NewGenesisBlock(dif uint64) *Block {
	return &Block{
		Header: &BlockHeader{
			ParentHash: []byte("0"),
			Time:       uint64(time.Now().Unix()),
			Difficulty: dif,
			Number:     big.NewInt(0),
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
