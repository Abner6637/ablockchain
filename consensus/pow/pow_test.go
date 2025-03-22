package pow

import (
	"ablockchain/core"
	"ablockchain/p2p"
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
			Difficulty: 2,
			Nonce:      0,
			Number:     13,
		},
	}
}

func newP2PNode() *p2p.Node {
	return &p2p.Node{}
}

func TestRun(t *testing.T) {
	block := newTestBlock()
	p2pnode := newP2PNode()
	pow := NewProofOfWork(p2pnode)
	pow.Run(block)

	if block.Header.Nonce == 0 {
		t.Errorf("Run() failed to find a valid nonce")
	}

	if len(block.Header.BlockHash()) != 32 {
		t.Errorf("Run() returned an invalid hash length: %d", len(block.Header.BlockHash()))
	}
}

// 测试 Validate 是否正确验证区块
func TestValidate(t *testing.T) {
	block := newTestBlock()
	p2pnode := newP2PNode()
	pow := NewProofOfWork(p2pnode)

	// 运行挖矿
	pow.Run(block)

	if !Validate(block) {
		t.Errorf("Validate() failed, block should be valid")
	}
}
