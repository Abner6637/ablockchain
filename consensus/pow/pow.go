package pow

import (
	"ablockchain/core"
	"ablockchain/crypto"
	"ablockchain/event"
	"ablockchain/p2p"
	"bytes"
	"fmt"
	"log"
	"math/big"
	"time"
)

type ProofOfWork struct {
	p2pNode *p2p.Node
	block   *core.Block
	target  *big.Int //用于判断hash前置0个数是否达到Difficulty要求
	running bool
}

func NewProofOfWork(p2pNode *p2p.Node) *ProofOfWork {
	return &ProofOfWork{
		p2pNode: p2pNode,
		running: false,
		block:   nil,
		target:  nil,
	}
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
func (pow *ProofOfWork) Run(block *core.Block) {
	var hashInt big.Int
	var hash []byte
	nonce := uint64(0)
	maxNonce := uint64(1000000)
	pow.block = block
	target := big.NewInt(1)
	target.Lsh(target, uint(256-pow.block.Header.Difficulty*4))
	pow.target = target

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

	block.Hash = hash[:]
	block.Header.Nonce = nonce
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
func (pow *ProofOfWork) Start() error {
	if pow.running {
		fmt.Println("PoW 已经在运行")
		return nil
	}

	pow.running = true
	go pow.ListenForConsensus()
	return nil
}

func (pow *ProofOfWork) Stop() error {
	if !pow.running {
		fmt.Println("PoW 已经停止")
		return nil
	}
	pow.running = false
	event.Bus.Publish("ConsensusStop", true)
	fmt.Println("PoW stop")
	return nil
}

// 监听共识事件
func (pow *ProofOfWork) ListenForConsensus() {
	consensusstart := event.Bus.Subscribe("ConsensusStart")
	consensusstop := event.Bus.Subscribe("ConsensusStop")
	for {
		select {
		case msg := <-consensusstart:
			block, ok := msg.(*core.Block)
			if !ok {
				log.Fatal("转换失败: 事件数据不是 *core.Block 类型")
			}
			fmt.Println("PoW 收到共识事件，开始计算区块:", block.Header.Number)
			pow.Run(block)
			if pow.Validate(block) {
				fmt.Println("验证通过，准备上链")
				event.Bus.Publish("ConsensusFinish", block)
			}
		case msg := <-consensusstop:
			if msg == true {
				fmt.Println("\n结束监听")
				return
			}
		default:
			time.Sleep(500 * time.Millisecond) // 避免 CPU 高占用
		}
	}
}
