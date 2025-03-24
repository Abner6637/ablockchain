package pbftcore

import (
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"ablockchain/event"
	"ablockchain/p2p"
	"log"
)

const (
	ByzantineSize int = 1
)

type Core struct {
	p2pNode *p2p.Node

	consensusState *consensusState

	events []event.EventSubscription
}

func (c *Core) Start() error {

	c.SubcribeEvents()

	c.HandleEvents()

	return nil
}

func (c *Core) Stop() error {
	event.Bus.Publish("ConsensusStop", true)
	log.Println("PBFT stop")
	return nil
}

func NewCore(p2pNode *p2p.Node) *Core {
	return &Core{
		p2pNode: p2pNode,
	}
}

func (c *Core) Broadcast(msg *pbfttypes.Message) error {
	payload, err := msg.EncodeMessage()
	if err != nil {
		return err
	}

	p2pMsg := &p2p.Message{
		Type: p2p.ConsensusMessage,
		Data: payload,
	}
	encodedP2PMsg, err := p2pMsg.Encode()
	if err != nil {
		return err
	}

	c.p2pNode.BroadcastMessage(string(encodedP2PMsg))

	return nil
}
