package core

import (
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/rlp"
)

const (
	SnapshotCall uint64 = iota
	SnapshotReply
)

type Snapshot struct {
	ID     string
	Number *big.Int // 快照最高区块的编号
	Hash   []byte   // 快照最高区块的哈希
	//Root   []byte   // StateRoot
}

func (s *Snapshot) EncodeSnapshot() ([]byte, error) {
	encodedSnapshot, err := rlp.EncodeToBytes(s)
	if err != nil {
		log.Fatal("RLP encoding failed:", err)
		return nil, err
	}
	return encodedSnapshot, nil
}

func DecodeSnapshot(data []byte) (*Snapshot, error) {
	var s Snapshot
	err := rlp.DecodeBytes(data, &s)
	if err != nil {
		log.Fatal("Failed to decode RLP data:", err)
		return nil, err
	}
	return &s, nil
}

func (bc *Blockchain) CreateSnapshot() *Snapshot {
	snapshot := &Snapshot{
		Number: new(big.Int).Set(bc.CurBlockNum),
		Hash:   bc.CurrentBlockHash,
		//Root:   bc.StateDB.StateRoot,
	}
	return snapshot
}

// TODO：恢复到某个快照状态
func (bc *Blockchain) RestoreSnapshot(snap *Snapshot) error {
	if snap == nil {
		return fmt.Errorf("nil snapshot")
	}

	bc.CurBlockNum = new(big.Int).Set(snap.Number)
	bc.CurrentBlockHash = snap.Hash
	//bc.StateRoot = snap.Root

	/* if err := bc.StateDB.RevertToRoot(snap.Root); err != nil {
		return fmt.Errorf("failed to revert state: %v", err)
	} */

	bc.TxPool.Clear()

	return nil
}

// 同步其他节点的快照
// TODO：同步快照（那就需要从其他节点获取StateDB以及区块等更多数据？）
func (bc *Blockchain) SyncSnapshot(snap *Snapshot) error {
	if snap == nil {
		return fmt.Errorf("nil snapshot")
	}

	bc.CurBlockNum = new(big.Int).Set(snap.Number)
	bc.CurrentBlockHash = snap.Hash
	//bc.StateRoot = snap.Root

	/* if err := bc.StateDB.RevertToRoot(snap.Root); err != nil {
		return fmt.Errorf("failed to revert state: %v", err)
	} */

	bc.TxPool.Clear()

	return nil
}

func (snap *Snapshot) PrintSnapshot() {
	if snap == nil {
		log.Println("Snapshot is nil")
		return
	}
	fmt.Printf("Snapshot ID: %s\n", snap.ID)
	fmt.Printf("Block Number: %s\n", snap.Number.String())
	fmt.Printf("Block Hash: %x\n", snap.Hash)
	//fmt.Printf("State Root: %x\n", snap.Root)
}

// type SnapshotMessage struct {
// 	Type uint64
// 	Data []byte
// }

// // 解码SnapshotMessage.Data
// func (m *SnapshotMessage) Decode(val interface{}) error {
// 	return rlp.DecodeBytes(m.Data, val)
// }

// func Encode(val interface{}) ([]byte, error) {
// 	return rlp.EncodeToBytes(val)
// }
