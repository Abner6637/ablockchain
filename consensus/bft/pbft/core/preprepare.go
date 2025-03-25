package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
)

func (c *Core) HandlePreprepare(msg *pbfttypes.Message) error {
	var preprepare *bft.Preprepare
	err := msg.Decode(&preprepare) // 直接针对msg.Msg进行解码
	if err != nil {
		return err
	}

	// TODO 验证阶段
	c.consensusState.setPreprepare(preprepare)
	c.setState(pbfttypes.StatePreprepared)

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

	// 只有主节点会发送Preprepare，主节点在SendPreprepare中存储Preprepare，其他节点在HandlePreprepare
	c.consensusState.setPreprepare(&bft.Preprepare{
		View:     c.consensusState.getView(),
		Sequence: c.consensusState.getSequence(),
		Request:  *request,
	})
	c.setState(pbfttypes.StatePreprepared)

	return nil
}
