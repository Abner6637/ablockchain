package core

import (
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/rlp"
)

type GetBlocksMessage struct {
	PeerID string
	From   *big.Int
	To     *big.Int
}

func (s *GetBlocksMessage) EncodeGetBlocksMessage() ([]byte, error) {
	encodedGetBlocksMessage, err := rlp.EncodeToBytes(s)
	if err != nil {
		log.Fatal("RLP encoding failed:", err)
		return nil, err
	}
	return encodedGetBlocksMessage, nil
}

func DecodeGetBlocksMessage(data []byte) (*GetBlocksMessage, error) {
	var s GetBlocksMessage
	err := rlp.DecodeBytes(data, &s)
	if err != nil {
		log.Fatal("Failed to decode RLP data:", err)
		return nil, err
	}
	return &s, nil
}

type BlocksMessage struct {
	PeerID string
	Blocks []*Block
}

func (s *BlocksMessage) EncodeBlocksMessage() ([]byte, error) {
	encodedBlocksMessage, err := rlp.EncodeToBytes(s)
	if err != nil {
		log.Fatal("RLP encoding failed:", err)
		return nil, err
	}
	return encodedBlocksMessage, nil
}

func DecodeBlocksMessage(data []byte) (*BlocksMessage, error) {
	var s BlocksMessage
	err := rlp.DecodeBytes(data, &s)
	if err != nil {
		log.Fatal("Failed to decode RLP data:", err)
		return nil, err
	}
	return &s, nil
}

// type GetSnapshotMessage struct {
// 	PeerID string
// }

type SnapshotMessage struct {
	PeerID   string
	Snapshot *Snapshot
}

func (s *SnapshotMessage) EncodeSnapshotMessage() ([]byte, error) {
	encodedSnapshotMessage, err := rlp.EncodeToBytes(s)
	if err != nil {
		log.Fatal("RLP encoding failed:", err)
		return nil, err
	}
	return encodedSnapshotMessage, nil
}

func DecodeSnapshotMessage(data []byte) (*SnapshotMessage, error) {
	var s SnapshotMessage
	err := rlp.DecodeBytes(data, &s)
	if err != nil {
		log.Fatal("Failed to decode RLP data:", err)
		return nil, err
	}
	return &s, nil
}
