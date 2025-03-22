package core

import (
	"ablockchain/crypto"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/libp2p/go-libp2p/core/network"
)

// rlp编码不支持time.Time，考虑改为Unix？
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

func (tx *Transaction) PrintTransaction() {
	if tx == nil {
		fmt.Println("Transaction is nil")
		return
	}

	// 打印 Transaction 信息
	fmt.Println("Transaction:")
	fmt.Printf("  Hash: %x\n", tx.Hash) // 输出字节数组为十六进制
	fmt.Printf("  From: %s\n", tx.From)
	fmt.Printf("  To: %s\n", tx.To)
	fmt.Printf("  Value: %d\n", tx.Value)
	fmt.Printf("  Gas: %d\n", tx.Gas)
	fmt.Printf("  Time: %v\n", tx.Time)
}
