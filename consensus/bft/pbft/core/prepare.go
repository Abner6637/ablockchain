package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
)

func (c *Core) HandlePrepare(msg *pbfttypes.Message) error {
	var prepare *bft.Prepare
	err := msg.Decode(&prepare)
	if err != nil {
		return err
	}

	c.consensusState.addPrepare(msg)

	if len(c.consensusState.Prepares.messages) >= 2*ByzantineSize+1 {
		c.consensusState.setState(pbfttypes.StatePrepared)
		c.SendCommit()
	}

	return nil
}

func (c *Core) SendPrepare() error {
	var msg pbfttypes.Message
	msg.Code = pbfttypes.MsgPrepare
	prepare, err := pbfttypes.Encode(c.consensusState.getPrepare())
	if err != nil {
		return err
	}
	msg.Msg = prepare

	c.Broadcast(&msg)

	return nil
}
