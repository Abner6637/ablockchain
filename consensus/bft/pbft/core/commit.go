package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"ablockchain/event"
	"errors"
)

func (c *Core) HandleCommit(msg *pbfttypes.Message) error {
	var commit *bft.Commit
	err := msg.Decode(&commit)
	if err != nil {
		return err
	}

	c.consensusState.addCommit(msg)

	if len(c.consensusState.Commits.messages) >= 2*ByzantineSize+1 {
		c.consensusState.setState(pbfttypes.StateCommitted)

		// TODO: 区块是从这时候获取的吗？还是从HandleRequest那里先在core里存一个block？
		block, err := c.consensusState.getBlock()
		if err != nil {
			return errors.New("invalid block")
		}
		event.Bus.Publish("ConsensusFinish", block)
	}

	return nil
}

func (c *Core) SendCommit() error {
	var msg pbfttypes.Message
	msg.Code = pbfttypes.MsgCommit
	commit, err := pbfttypes.Encode(c.consensusState.getCommit())
	if err != nil {
		return err
	}
	msg.Msg = commit

	c.Broadcast(&msg)

	return nil
}
