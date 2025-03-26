package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"log"
)

func (c *Core) HandlePreprepare(msg *pbfttypes.Message) error {
	// 验证消息签名
	if err := VerifySignature(msg); err != nil {
		return err
	}

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
	msg.Code = pbfttypes.MsgPreprepare
	preprepare, err := pbfttypes.Encode(&bft.Preprepare{
		View:     c.consensusState.getView(),
		Sequence: c.consensusState.getSequence(),
		Request:  *request,
	})

	// 打印使用
	log.Printf("生成Preprepare：%+v", bft.Preprepare{
		View:     c.consensusState.getView(),
		Sequence: c.consensusState.getSequence(),
		Request:  *request,
	})

	if err != nil {
		return err
	}
	msg.Msg = preprepare
	msg.Address = c.address
	msg.Signature, err = c.SignMessage(&msg)
	if err != nil {
		return err
	}

	log.Printf("广播Preprepare消息：%+v", msg)

	c.Broadcast(&msg)

	// 只有主节点会发送Preprepare，
	// 主节点在SendPreprepare中存储Preprepare，
	// 其他节点在HandlePreprepare中存储Preprepare
	c.consensusState.setPreprepare(&bft.Preprepare{
		View:     c.consensusState.getView(),
		Sequence: c.consensusState.getSequence(),
		Request:  *request,
	})
	c.setState(pbfttypes.StatePreprepared)

	return nil
}
