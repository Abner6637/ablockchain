package pbftcore

import (
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"ablockchain/core"
	"ablockchain/event"
	"ablockchain/p2p"
	"log"
	"math/big"
)

type Core struct {
	p2pNode *p2p.Node // 打算用p2pNode的ID标识共识节点的地址

	consensusState *consensusState

	state            pbfttypes.State
	curCommitedBlock *core.Block

	events []event.EventSubscription

	Primary       string
	NodeSet       []string // 通过config注入
	ByzantineSize int
}

func NewCore(p2pNode *p2p.Node) *Core {
	return &Core{
		p2pNode: p2pNode,
		state:   pbfttypes.StateAcceptRequest,
	}
}

func (c *Core) Start() error {

	c.SubcribeEvents()

	c.HandleEvents()

	c.StartNewProcess(big.NewInt(0))

	return nil
}

func (c *Core) Stop() error {
	event.Bus.Publish("ConsensusStop", true)
	log.Println("PBFT stop")
	return nil
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

func (c *Core) IsPrimary() bool {
	return c.Primary == c.p2pNode.ID
}

func (c *Core) setState(state pbfttypes.State) {
	c.state = state
}

func (c *Core) StartNewProcess(num *big.Int) {
	if c.consensusState == nil {
		NewConsensusState(big.NewInt(0), big.NewInt(0), nil)
		log.Printf("Initiate the consensus state\n")
	} else {
		NewConsensusState(c.consensusState.getView(), big.NewInt(int64(c.curCommitedBlock.Header.Number)+1), nil)
	}

}
