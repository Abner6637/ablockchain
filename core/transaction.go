package core

import (
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

func encodeTransaction(tx Transaction) ([]byte, error) {
	encodedTx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		log.Fatal("RLP encoding failed:", err)
		return nil, err
	}
	return encodedTx, nil
}

func receiveData(stream network.Stream) {
	buf := make([]byte, 1024)
	n, err := stream.Read(buf)
	if err != nil {
		log.Fatal("Failed to read from stream:", err)
	}

	// 解码 RLP 数据
	var tx Transaction
	err = rlp.DecodeBytes(buf[:n], &tx)
	if err != nil {
		log.Fatal("Failed to decode RLP data:", err)
	}

	log.Printf("Received transaction: %+v", tx)
}
