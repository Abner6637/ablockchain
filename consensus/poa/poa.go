package poa

import (
	"ablockchain/p2p"
)

type ProofOfAuthority struct {
	p2pNode *p2p.Node
	running bool
}

func NewProofOfAuthority(p2pNode *p2p.Node) *ProofOfAuthority {
	return &ProofOfAuthority{
		p2pNode: p2pNode,
		running: false,
	}
}

func (poa *ProofOfAuthority) Start() error {
	poa.running = true
	return nil
}

func (poa *ProofOfAuthority) Stop() error {
	poa.running = false
	return nil
}
