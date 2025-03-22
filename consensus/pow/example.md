```go
package main

import (
	"ablockchain/consensus"
	"ablockchain/consensus/pow"
	"ablockchain/core"
	"ablockchain/p2p"
	"ablockchain/system"
	"fmt"
	"time"
)

func main() {
	var consensus consensus.Consensus
	var err error

	p2pNode := newP2PNode()
	if p2pNode == nil {
		fmt.Println("Failed to create P2P node, exiting...")
		return
	}

	consensus = pow.NewProofOfWork(p2pNode)

	blockchain, err := core.NewBlockchain()
	if err != nil || blockchain == nil {
		fmt.Println("Failed to initialize blockchain, exiting...")
		return
	}

	go system.ListenNewBlocks(blockchain)
	go consensus.Start()
	time.Sleep(1 * time.Second)


	if blockchain.NewBlockChan == nil {
		fmt.Println("Error: blockchain.NewBlockChan is nil")
		return
	}

	blockchain.NewBlockChan <- newTestBlock()

	time.Sleep(10 * time.Second)
	consensus.Stop()

	time.Sleep(2 * time.Second)
}

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
	node, err := p2p.NewNode("/ip4/127.0.0.1/tcp/0")
	if err != nil {
		fmt.Println("Failed to create P2P node:", err)
		return nil
	}
	return node
}

```

测试中断操作
```go
package main

import (
	"ablockchain/consensus"
	"ablockchain/consensus/pow"
	"ablockchain/core"
	"ablockchain/event"
	"ablockchain/p2p"
	"ablockchain/system"
	"encoding/hex"
	"fmt"
	"log"
	"time"
)

func main() {
	var consensus consensus.Consensus
	var err error

	p2pNode := newP2PNode()
	if p2pNode == nil {
		fmt.Println("Failed to create P2P node, exiting...")
		return
	}

	consensus = pow.NewProofOfWork(p2pNode)

	blockchain, err := core.NewBlockchain()
	if err != nil || blockchain == nil {
		fmt.Println("Failed to initialize blockchain, exiting...")
		return
	}

	go system.ListenNewBlocks(blockchain)
	go consensus.Start()
	time.Sleep(2 * time.Second)

	if blockchain.NewBlockChan == nil {
		fmt.Println("Error: blockchain.NewBlockChan is nil")
		return
	}
	//blockchain.NewBlockChan <- newTestBlock() //模拟打包区块
	event.Bus.Publish("ConsensusStart", newTestBlock())
	time.Sleep(1 * time.Second)
	event.Bus.Publish("BlockMessage", newValidBlock())

	time.Sleep(30 * time.Second)
	consensus.Stop()

	time.Sleep(2 * time.Second)
}

func newTestBlock() *core.Block {
	// 将十六进制字符串转换为 []byte
	blockHash, err := hex.DecodeString("00000ccfbb0d507f9e39705ea1d8bc8b28774755dd87e050c890f515a4bf6641")
	if err != nil {
		log.Fatal("BlockHash 解析失败:", err)
	}

	parentHash, err := hex.DecodeString("0cee80f09bae44967f99f068c38ca265d83d00d31b3f86db9bd502772cc8e781")
	if err != nil {
		log.Fatal("ParentHash 解析失败:", err)
	}

	merkleRoot, err := hex.DecodeString("0cee80f09bae44967f99f068c38ca265d83d00d31b3f86db9bd502772cc8e781")
	if err != nil {
		log.Fatal("MerkleRoot 解析失败:", err)
	}

	// 构造区块
	return &core.Block{
		Hash: blockHash,
		Header: &core.BlockHeader{
			ParentHash: parentHash,
			Time:       time.Date(2025, 3, 22, 8, 8, 22, 105293644, time.UTC),
			Difficulty: 6,
			Number:     13,
			MerkleRoot: merkleRoot,
			Nonce:      216994,
		},
	}
}

func newValidBlock() *core.Block {
	// 将十六进制字符串转换为 []byte
	blockHash, err := hex.DecodeString("00000ccfbb0d507f9e39705ea1d8bc8b28774755dd87e050c890f515a4bf6641")
	if err != nil {
		log.Fatal("BlockHash 解析失败:", err)
	}

	parentHash, err := hex.DecodeString("0cee80f09bae44967f99f068c38ca265d83d00d31b3f86db9bd502772cc8e781")
	if err != nil {
		log.Fatal("ParentHash 解析失败:", err)
	}

	merkleRoot, err := hex.DecodeString("0cee80f09bae44967f99f068c38ca265d83d00d31b3f86db9bd502772cc8e781")
	if err != nil {
		log.Fatal("MerkleRoot 解析失败:", err)
	}

	// 构造区块
	return &core.Block{
		Hash: blockHash,
		Header: &core.BlockHeader{
			ParentHash: parentHash,
			Time:       time.Date(2025, 3, 22, 8, 8, 22, 105293644, time.UTC),
			Difficulty: 5,
			Number:     13,
			MerkleRoot: merkleRoot,
			Nonce:      216994,
		},
	}
}

func newP2PNode() *p2p.Node {
	node, err := p2p.NewNode("/ip4/127.0.0.1/tcp/0")
	if err != nil {
		fmt.Println("Failed to create P2P node:", err)
		return nil
	}
	return node
}

```