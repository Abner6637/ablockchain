package pow

import (
	"ablockchain/core"
	"fmt"
	"testing"
	"time"
)

// 创建一个简单的测试区块头
func createTestBlockHeader() *core.BlockHeader {
	return &core.BlockHeader{
		ParentHash: []byte("0df9a8f4a2f2fc354c3c8aa5e837d4db137f20ccbf3d8336e4c95ac9d0e2943e"),
		Time:       time.Now(),
		Difficulty: 1,
		MerkleRoot: []byte("0df9a8f4a2f2fc354c3c8aa5e837d4db137f20ccbf3d8336e4c95ac9d0e2943e"),
		Nonce:      0,
	}
}

// 测试 NewPoWProof 函数
func TestNewPoWProof(t *testing.T) {
	blockHeader := createTestBlockHeader()
	powProof := NewPoWProof(blockHeader)

	if powProof.BlockHeader == nil {
		t.Errorf("Failed to initialize PowProof correctly")
	}
}

// 测试 calculateHash 方法
func TestCalculateHash(t *testing.T) {
	blockHeader := createTestBlockHeader()
	powProof := NewPoWProof(blockHeader)

	// 测试特定 nonce 的哈希计算
	nonce := uint32(0)
	hash := powProof.calculateHash(nonce)

	if len(hash) == 0 {
		t.Errorf("calculateHash() failed, expected non-empty hash")
	}

	// 输出计算的哈希以便查看
	t.Logf("Calculated hash: %x", hash)
}

// 测试 mine 方法
func TestMine(t *testing.T) {
	blockHeader := createTestBlockHeader()
	powProof := NewPoWProof(blockHeader)

	// 测试 PoW 持续计算直到找到满足条件的 nonce
	nonce := powProof.mine()
	fmt.Print(powProof.GetBlockHash())

	if nonce == 0 {
		t.Errorf("mine() failed, expected nonce to be non-zero")
	}

	// 输出找到的 nonce
	t.Logf("Found nonce: %d", nonce)
}

// 测试 GetBlockHash 方法
func TestGetBlockHash(t *testing.T) {
	blockHeader := createTestBlockHeader()
	powProof := NewPoWProof(blockHeader)

	hash := powProof.GetBlockHash()

	if len(hash) == 0 {
		t.Errorf("GetBlockHash() failed, expected non-empty hash")
	}

	// 输出最终计算的哈希值
	t.Logf("Final Block Hash: %x", hash)
}

// 测试 Start 方法
func TestStart(t *testing.T) {
	blockHeader := createTestBlockHeader()
	powProof := NewPoWProof(blockHeader)

	block := &core.Block{Header: blockHeader}

	// 测试 Start 方法
	err := powProof.Start(block)
	if err != nil {
		t.Errorf("Start() failed: %v", err)
	}
}

// 测试 Stop 方法
func TestStop(t *testing.T) {
	blockHeader := createTestBlockHeader()
	powProof := NewPoWProof(blockHeader)

	err := powProof.Stop()
	if err != nil {
		t.Errorf("Stop() failed: %v", err)
	}
}
