package hotstuffcore

import (
	"ablockchain/consensus/bft"
	hotstufftypes "ablockchain/consensus/bft/hotstuff/types"
	"log"
)

func (c *Core) HandlePrepare2(msg *hotstufftypes.Message) error {
	if c.IsPrimary() {
		log.Printf("主节点，丢弃Prepare2消息：%+v", msg)
		return nil
	}

	log.Printf("开始处理Prepare2消息：%+v", msg)

	// 验证聚合签名
	if err := c.VerifyAggregateSignature(msg); err != nil {
		return err
	}

	var prepare *bft.Prepare
	err := msg.Decode(&prepare)
	if err != nil {
		return err
	}

	if c.consensusState.Preprepare != nil {
		if prepare.View.Cmp(c.consensusState.View) != 0 || prepare.Sequence.Cmp(c.consensusState.Sequence) != 0 {
			log.Printf("Prepare2消息的View或Sequence不匹配，丢弃消息")
			return nil
		}
	} else {
		return nil
	}

	c.consensusState.addPrepare(msg)
	log.Printf("目前已存储的Prepare消息数目：%d", len(c.consensusState.Prepares.messages))

	if c.state.Cmp(hotstufftypes.StatePrepared) < 0 {
		c.setState(hotstufftypes.StatePrepared)
		c.SendCommit1()
	}

	return nil
}

func (c *Core) SendPrepare2() error {
	var msg hotstufftypes.Message
	msg.Code = hotstufftypes.MsgPrepare2

	prepare := c.consensusState.getPrepare()

	encodedPrepare, err := hotstufftypes.Encode(prepare)
	if err != nil {
		return err
	}

	// 打印使用
	log.Printf("生成Prepare2：%+v", prepare)

	msg.Msg = encodedPrepare
	msg.Address = c.address

	aggSig, valList, err := c.SignAggregate(c.consensusState.Prepares)
	if err != nil {
		log.Printf("聚合签名失败：%v", err)
	}

	msg.Additional = valList
	msg.Signature = aggSig

	if err != nil {
		return err
	}

	// 发给别人Prepare消息时，自己也保存一份自己发送的
	c.consensusState.addPrepare(&msg)

	log.Printf("广播Preppare2消息：%+v", msg)
	c.Broadcast(&msg)

	return nil
}
