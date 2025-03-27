package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"log"
)

func (c *Core) HandlePreprepare(msg *pbfttypes.Message) error {
	log.Printf("开始处理Preprepare消息：%+v", msg)

	// 验证消息签名
	if err := VerifySignature(msg); err != nil {
		return err
	}

	var preprepare *bft.Preprepare
	err := msg.Decode(&preprepare) // 直接针对msg.Msg进行解码
	if err != nil {
		return err
	}
	log.Printf("解码后得到的Preprepare：%+v", preprepare)

	if preprepare.View.Cmp(c.consensusState.View) != 0 || preprepare.Sequence.Cmp(c.consensusState.Sequence) != 0 {
		log.Printf("警告：接收到的preprepare的view或sequence不匹配")
		return nil
	}

	if c.state.Cmp(pbfttypes.StatePreprepared) >= 0 {
		log.Printf("警告：当前core的状态>=StatePreprepared")
		return nil
	}

	c.consensusState.setPreprepare(preprepare)
	c.setState(pbfttypes.StatePreprepared)

	c.SendPrepare()

	return nil
}

func (c *Core) SendPreprepare(request *bft.Request) error {
	var msg pbfttypes.Message
	msg.Code = pbfttypes.MsgPreprepare

	preprepare := &bft.Preprepare{
		View:     c.consensusState.getView(),
		Sequence: c.consensusState.getSequence(),
		Request:  *request,
	}

	encodedPreprepare, err := pbfttypes.Encode(preprepare)

	// 打印使用
	log.Printf("生成Preprepare：%+v", preprepare)

	if err != nil {
		return err
	}
	msg.Msg = encodedPreprepare
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
