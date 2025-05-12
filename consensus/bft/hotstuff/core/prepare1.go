package hotstuffcore

import (
	"ablockchain/consensus/bft"
	hotstufftypes "ablockchain/consensus/bft/hotstuff/types"
	"log"
)

func (c *Core) HandlePrepare1(msg *hotstufftypes.Message) error {
	if !c.IsPrimary() {
		log.Printf("非主节点，丢弃Prepare1消息：%+v", msg)
		return nil
	}

	log.Printf("开始处理Prepare1消息：%+v", msg)

	// 验证消息签名
	if err := c.VerifySignature(msg); err != nil {
		return err
	}

	var prepare *bft.Prepare
	err := msg.Decode(&prepare)
	if err != nil {
		return err
	}

	if c.consensusState.Preprepare != nil {
		if prepare.View.Cmp(c.consensusState.View) != 0 {
			return nil
		}
	} else {
		return nil
	}

	c.consensusState.addPrepare(msg)
	log.Printf("目前已存储的Prepare消息数目：%d", len(c.consensusState.Prepares.messages))

	// 2f个即可，因为还有一个是Preprepare
	if len(c.consensusState.Prepares.messages) >= 2*c.ByzantineSize && c.state.Cmp(hotstufftypes.StatePrepared) < 0 {
		c.setState(hotstufftypes.StatePrepared)

		c.SendPrepare2()
	}

	return nil
}

func (c *Core) SendPrepare1() error {
	var msg hotstufftypes.Message
	msg.Code = hotstufftypes.MsgPrepare1

	prepare := c.consensusState.getPrepare()

	encodedPrepare, err := hotstufftypes.Encode(prepare)
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

	log.Printf("发送Prepare1消息给主节点：%+v", msg)
	// c.Broadcast(&msg)
	c.Unicast(&msg, string(c.Primary))

	return nil
}
