package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"log"
)

func (c *Core) HandlePrepare(msg *pbfttypes.Message) error {
	log.Printf("开始处理Prepare消息：%+v", msg)

	// 验证消息签名
	if err := VerifySignature(msg); err != nil {
		return err
	}

	var prepare *bft.Prepare
	err := msg.Decode(&prepare)
	if err != nil {
		return err
	}

	if c.consensusState.Preprepare != nil {
		if prepare.View.Cmp(c.consensusState.View) != 0 || prepare.Sequence.Cmp(c.consensusState.Sequence) != 0 {
			return nil
		}
	} else {
		return nil
	}

	c.consensusState.addPrepare(msg)
	log.Printf("目前已存储的Prepare消息数目：%d", len(c.consensusState.Prepares.messages))

	// 2f个即可，因为还有一个是Preprepare
	if len(c.consensusState.Prepares.messages) >= 2*c.ByzantineSize && c.state.Cmp(pbfttypes.StatePrepared) < 0 {
		c.setState(pbfttypes.StatePrepared)
		c.SendCommit()
	}

	return nil
}

func (c *Core) SendPrepare() error {
	var msg pbfttypes.Message
	msg.Code = pbfttypes.MsgPrepare

	prepare := c.consensusState.getPrepare()

	encodedPrepare, err := pbfttypes.Encode(prepare)
	if err != nil {
		return err
	}

	// 打印使用
	log.Printf("生成Prepare：%+v", prepare)

	msg.Msg = encodedPrepare
	msg.Address = c.address
	msg.Signature, err = c.SignMessage(&msg)
	if err != nil {
		return err
	}

	// 发给别人Prepare消息时，自己也保存一份自己发送的
	c.consensusState.addPrepare(&msg)

	log.Printf("广播Preppare消息：%+v", msg)
	c.Broadcast(&msg)

	return nil
}
