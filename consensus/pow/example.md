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