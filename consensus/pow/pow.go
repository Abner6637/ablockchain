package pow

import (
	"ablockchain/crypto"
	"encoding/hex"
	"fmt"

	"ablockchain/core"
)

type PowProof struct {
	BlockHeader *core.BlockHeader
}

func NewPoWProof(BlockHeader *core.BlockHeader) *PowProof {
	consensus := new(PowProof)
	consensus.BlockHeader = BlockHeader

	return consensus
}

// 实现 Start 方法
func (p *PowProof) Start(block *core.Block) error {
	fmt.Println("PoW 共识已启动")
	nounce := p.mine()
	fmt.Println(nounce)
	return nil
}

// 实现 Stop 方法
func (p *PowProof) Stop() error {
	fmt.Println("PoW 共识已停止")
	return nil
}

// 计算区块的哈希值
func (pow *PowProof) calculateHash(nonce uint32) []byte {
	// 拼接区块头数据并计算其哈希
	data := append(pow.BlockHeader.ParentHash, []byte(fmt.Sprintf("%v", pow.BlockHeader.Time))...)
	data = append(data, []byte(fmt.Sprintf("%d", pow.BlockHeader.Difficulty))...)
	data = append(data, pow.BlockHeader.MerkleRoot...)
	data = append(data, []byte(fmt.Sprintf("%d", nonce))...)

	hash := crypto.GlobalHashAlgorithm.Hash(data)
	return hash[:]
}

// 判断哈希是否符合难度要求
func (pow *PowProof) isValidHash(hash []byte) bool {
	// 根据 Difficulty 生成目标难度
	target := make([]byte, len(hash))
	// 生成目标值，即哈希值需要的前导零数量
	for i := uint64(0); i < pow.BlockHeader.Difficulty; i++ {
		target[i/8] |= 1 << (7 - (i % 8))
	}
	// 如果哈希小于目标值，说明符合难度要求
	return compareHashes(hash, target) < 0
}

// 比较两个哈希值，返回第一个小于还是大于第二个
func compareHashes(a, b []byte) int {
	for i := 0; i < len(a); i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

// 进行工作量证明，返回找到的Nonce
func (pow *PowProof) mine() uint32 {
	var nonce uint32
	for {
		hash := pow.calculateHash(nonce)
		if pow.isValidHash(hash) {
			return nonce
		}
		nonce++
	}
}

// 获取区块的哈希值
func (pow *PowProof) GetBlockHash() []byte {
	nonce := pow.mine()
	hash := pow.calculateHash(nonce)
	return hash
}

// 辅助函数：显示区块哈希（用于调试）
func (pow *PowProof) DisplayBlockHash() {
	hash := pow.GetBlockHash()
	fmt.Printf("Block Hash: %s\n", hex.EncodeToString(hash))
}
