package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
)

func (c *Core) HandlePreprepare(msg *pbfttypes.Message) error {
	var preprepare *bft.Preprepare
	err := msg.Decode(&preprepare)
	if err != nil {
		return err
	}

	c.AcceptPreprepare(preprepare)
	c.SendPrepare()

	return nil
}

func (c *Core) SendPreprepare(request *bft.Request) error {
	var msg pbfttypes.Message
	msg.Code = pbfttypes.MsgPrepare
	preprepare, err := pbfttypes.Encode(&bft.Preprepare{
		View:     c.consensusState.getView(),
		Sequence: c.consensusState.getSequence(),
		Request:  *request,
	})
	if err != nil {
		return err
	}
	msg.Msg = preprepare

	c.Broadcast(&msg)

	return nil
}

func (c *Core) AcceptPreprepare(preprepare *bft.Preprepare) {
	c.consensusState.state = pbfttypes.StatePrepared
}
