package pow

import (
	"ablockchain/core"
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

// 创建一个测试区块
func newTestBlock() *core.Block {
	return &core.Block{
		Header: &core.BlockHeader{
			ParentHash: []byte("0df9a8f4a2f2fc354c3c8aa5e837d4db137f20ccbf3d8336e4c95ac9d0e2943e"),
			MerkleRoot: []byte("1cdfdf5680f2a639732f6aae64a8b96c10a913b46c8fcd908c9eb95925979974"),
			Time:       time.Now(),
			Difficulty: 3,
			Nonce:      0,
			Number:     13,
		},
	}
}

// 测试 prepareData 生成的数据格式
func TestPrepareData(t *testing.T) {
	block := newTestBlock()
	pow := NewProofOfWork(block)

	nonce := uint64(12345)
	data := pow.prepareData(nonce)

	expected := append(block.Header.ParentHash, block.Header.MerkleRoot...)
	expected = append(expected, []byte(hex.EncodeToString([]byte(fmt.Sprintf("%v", block.Header.Time))))...)
	expected = append(expected, []byte(fmt.Sprintf("%d", block.Header.Difficulty))...)
	expected = append(expected, []byte(fmt.Sprintf("%d", nonce))...)

	if !bytes.Contains(data, block.Header.ParentHash) || !bytes.Contains(data, block.Header.MerkleRoot) {
		t.Errorf("prepareData() missing necessary components")
	}
}

// 测试 Run 计算 nonce 是否成功
func TestRun(t *testing.T) {
	block := newTestBlock()
	pow := NewProofOfWork(block)

	nonce, hash := pow.Run()

	if nonce == 0 {
		t.Errorf("Run() failed to find a valid nonce")
	}

	if len(hash) != 32 {
		t.Errorf("Run() returned an invalid hash length: %d", len(hash))
	}
}

// 测试 Validate 是否正确验证区块
func TestValidate(t *testing.T) {
	block := newTestBlock()
	pow := NewProofOfWork(block)

	// 运行挖矿
	nonce, hash := pow.Run()
	block.Header.Nonce = nonce
	block.Hash = hash

	if !pow.Validate(block) {
		t.Errorf("Validate() failed, block should be valid")
	}
}
