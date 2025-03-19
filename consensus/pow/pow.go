package pow

import (
	"ablockchain/core"
	"ablockchain/crypto"
	"bytes"
	"fmt"
	"math/big"
)

type ProofOfWork struct {
	block  *core.Block
	target *big.Int //用于判断hash前置0个数是否达到Difficulty要求
}

func NewProofOfWork(b *core.Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-b.Header.Difficulty*4))

	pow := &ProofOfWork{b, target}

	return pow
}

// 准备用于计算hash的数据
func (pow *ProofOfWork) prepareData(nonce uint64) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.Header.ParentHash,
			pow.block.Header.MerkleRoot,
			[]byte(fmt.Sprintf("%v", pow.block.Header.Time)),
			[]byte(fmt.Sprintf("%d", pow.block.Header.Difficulty)),
			[]byte(fmt.Sprintf("%d", nonce)),
		},
		[]byte{},
	)
	return data
}

// 共识的核心逻辑
func (pow *ProofOfWork) Run() (uint64, []byte) {
	var hashInt big.Int
	var hash []byte
	nonce := uint64(0)
	maxNonce := uint64(1000000)

	fmt.Printf("Mining the block \"%d\"\n", pow.block.Header.Number)
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = crypto.GlobalHashAlgorithm.Hash(data)
		hashInt.SetBytes(hash[:])
		//结束条件：hash小于target
		if hashInt.Cmp(pow.target) == -1 {
			fmt.Printf("\n hash: %x", hash)
			fmt.Printf("\n nonce: %d", nonce)
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

// 挖矿完成后重新封装block
func NewBlock(block *core.Block) *core.Block {
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Header.Nonce = nonce

	return block
}

// 验证hash
func (pow *ProofOfWork) Validate(block *core.Block) bool {
	var hashInt big.Int

	data := pow.prepareData(block.Header.Nonce)
	hash := crypto.GlobalHashAlgorithm.Hash(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}

// 实现共识接口
func (pow *ProofOfWork) start() error {
	fmt.Printf("pow start")

	return nil
}
