package core

import (
	"ablockchain/crypto"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/libp2p/go-libp2p/core/network"
)

type Transaction struct {
	Time time.Time
	Hash []byte

	From  string
	To    string
	Value uint64
	Gas   uint64
}

func (tx *Transaction) EncodeTx() ([]byte, error) {
	encodedTx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		log.Fatal("RLP encoding failed:", err)
		return nil, err
	}
	return encodedTx, nil
}

// TODO:rlp解码的时候，传入的是数据地址还是数据本身？
// 涉及到解码的部分都需要修改
func DecodeTx(data []byte) (*Transaction, error) {
	var tx Transaction
	err := rlp.DecodeBytes(data, &tx)
	if err != nil {
		log.Fatal("Failed to decode RLP data:", err)
		return nil, err
	}
	return &tx, nil
}

func ReceiveData(stream network.Stream) *Transaction {
	buf := make([]byte, 1024)
	_, err := stream.Read(buf)
	if err != nil {
		log.Fatal("Failed to read from stream:", err)
	}

	tx, err := DecodeTx(buf)
	if err != nil {
		log.Fatal("Failed to decode RLP data:", err)
	}
	return tx

	// log.Printf("Received transaction: %+v", tx)
}

// 计算交易列表的 Merkle Root
func CalculateMerkleRoot(transactions []*Transaction) []byte {
	var txHashes [][]byte
	for _, tx := range transactions {
		txHashes = append(txHashes, tx.Hash)
	}
	return crypto.ComputeMerkleRoot(txHashes)
}
