package pbftcore

import (
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"errors"
)

// TODO: after events finished
func (c *Core) HandleEvents() {
	go func() {
		for {
			select {
			//case
			}
		}
	}()
}

func (c *Core) HandleMessage(payload []byte) error {
	msg, err := pbfttypes.DecodeMessage(payload)
	if err != nil {
		return errors.New("cannot decode message")
	}

	// TODO: different kinds of message
	switch msg.Code {
	case pbfttypes.MsgPreprepare:
		err := c.HandlePreprepare(msg)
		return err
	}

	return errors.New("invalid message")
}
