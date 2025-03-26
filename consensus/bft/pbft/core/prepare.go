package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
)

func (c *Core) HandlePrepare(msg *pbfttypes.Message) error {
	// 验证消息签名
	if err := VerifySignature(msg); err != nil {
		return err
	}

	var prepare *bft.Prepare
	err := msg.Decode(&prepare)
	if err != nil {
		return err
	}

	c.consensusState.addPrepare(msg)

	// 2f个即可，因为还有一个是Preprepare
	if len(c.consensusState.Prepares.messages) >= 2*c.ByzantineSize {
		c.setState(pbfttypes.StatePrepared)
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
	msg.Address = c.address
	msg.Signature, err = c.SignMessage(&msg)
	if err != nil {
		return err
	}

	// 发给别人Prepare消息时，自己也保存一份自己发送的
	c.consensusState.addPrepare(&msg)
	c.Broadcast(&msg)

	return nil
}
