package core

import (
	"ablockchain/crypto"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/libp2p/go-libp2p/core/network"
)

type Transaction struct {
	chainId uint64
	Time    uint64
	TxHash  []byte

	From  string
	To    string
	Value uint64
	Data  string
	Nonce uint64

	GasPrice uint64
	GasLimit uint64
	Gas      uint64
}

func (tx *Transaction) Encode() (any, any) {
	panic("unimplemented")
}

func (tx *Transaction) Decode(data []byte) any {
	panic("unimplemented")
}

func (tx *Transaction) DecodeTransaction(data []byte) any {
	panic("unimplemented")
}

func (tx *Transaction) EncodeTransaction() (any, any) {
	panic("unimplemented")
}

func (tx *Transaction) Hash() {
	panic("unimplemented")
}

func NewTransaction(from string, to string, value uint64, data string, nonce uint64) *Transaction {
	tx := &Transaction{
		From:  from,
		To:    to,
		Value: value,
		Data:  data,
		Nonce: nonce,
		Time:  uint64(time.Now().Unix()),
	}
	encodeTx, err := tx.EncodeTx()
	if err != nil {
		log.Fatal("EncodeTx failed:", err)
		return nil
	}

	hash := crypto.GlobalHashAlgorithm.Hash(encodeTx)
	tx.TxHash = hash[:]
	return tx
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

func (t *Transaction) MarshalJSON() ([]byte, error) {
	type Alias Transaction
	return json.Marshal(&struct {
		*Alias
		TxHash string `json:"TxHash"`
	}{
		Alias:  (*Alias)(t),
		TxHash: fmt.Sprintf("%x", t.TxHash), // 将哈希转换为十六进制字符串
	})
}

func (tx *Transaction) PrintTransaction() {
	if tx == nil {
		fmt.Println("Transaction is nil")
		return
	}
	// 打印 Transaction 信息
	fmt.Println("Transaction:")
	fmt.Printf("  Hash: %x\n", tx.TxHash) // 输出字节数组为十六进制
	fmt.Printf("  From: %s\n", tx.From)
	fmt.Printf("  To: %s\n", tx.To)
	fmt.Printf("  Value: %d\n", tx.Value)
	fmt.Printf("  Nonce: %d\n", tx.Nonce)
	fmt.Printf("  Time: %v\n", tx.Time)
}

func (tx *Transaction) VerifySignature(signature []byte) (bool, error) {
	encodedTx, err := tx.EncodeTx()
	if err != nil {
		return false, err
	}
	hashTx := crypto.GlobalHashAlgorithm.Hash(encodedTx)
	// 从哈希和签名恢复出公钥
	pubKey, err := crypto.SigToPub(hashTx, signature)
	if err != nil {
		return false, err
	}
	// 计算公钥对应的地址
	recoveredAddress := crypto.PubkeyToAddress(*pubKey).Hex()
	// 比对地址是否一致
	return recoveredAddress == tx.From, nil
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
