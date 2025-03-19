package pbftcore

import (
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"ablockchain/p2p"
)

type PBFT struct {
	p2pNode *p2p.Node
	view    uint64
	round   uint64
}

func (p *PBFT) Start() error {
	return nil
}

func (p *PBFT) Stop() error {
	return nil
}

func NewPBFT(p2pNode *p2p.Node) *PBFT {
	return &PBFT{
		p2pNode: p2pNode,
	}
}

func (p *PBFT) Broadcast(msg *pbfttypes.Message) error {
	payload, err := msg.EncodeMsg()
	if err != nil {
		return err
	}

	p.p2pNode.BroadcastMessage(string(payload))

	return nil
}

func (p *PBFT) HandleMessage(msg *pbfttypes.Message) {

}
