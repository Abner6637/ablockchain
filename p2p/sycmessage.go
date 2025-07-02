package p2p

import (
	"ablockchain/core"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/rlp"
)

type GetBlocksMsg struct {
	PeerID string
	From   *big.Int
	To     *big.Int
}

func (s *GetBlocksMsg) EncodeGetBlocksMessage() ([]byte, error) {
	encodedGetBlocksMessage, err := rlp.EncodeToBytes(s)
	if err != nil {
		log.Fatal("RLP encoding failed:", err)
		return nil, err
	}
	return encodedGetBlocksMessage, nil
}

func DecodeGetBlocksMessage(data []byte) (*GetBlocksMsg, error) {
	var s GetBlocksMsg
	err := rlp.DecodeBytes(data, &s)
	if err != nil {
		log.Fatal("Failed to decode RLP data:", err)
		return nil, err
	}
	return &s, nil
}

type BlocksMsg struct {
	PeerID string
	Blocks []*core.Block
}

func (s *BlocksMsg) EncodeBlocksMessage() ([]byte, error) {
	encodedBlocksMessage, err := rlp.EncodeToBytes(s)
	if err != nil {
		log.Fatal("RLP encoding failed:", err)
		return nil, err
	}
	return encodedBlocksMessage, nil
}

func DecodeBlocksMessage(data []byte) (*BlocksMsg, error) {
	var s BlocksMsg
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

type SnapshotMsg struct {
	PeerID   string
	Snapshot *core.Snapshot
}

func (s *SnapshotMsg) EncodeSnapshotMessage() ([]byte, error) {
	encodedSnapshotMessage, err := rlp.EncodeToBytes(s)
	if err != nil {
		log.Fatal("RLP encoding failed:", err)
		return nil, err
	}
	return encodedSnapshotMessage, nil
}

func DecodeSnapshotMessage(data []byte) (*SnapshotMsg, error) {
	var s SnapshotMsg
	err := rlp.DecodeBytes(data, &s)
	if err != nil {
		log.Fatal("Failed to decode RLP data:", err)
		return nil, err
	}
	return &s, nil
}
