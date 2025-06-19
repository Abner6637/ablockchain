package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"bytes"
	"log"
)

func (c *Core) HandlePreprepare(msg *pbfttypes.Message) error {
	log.Printf("开始处理Preprepare消息：%+v", msg)

	// 验证消息签名
	if err := VerifySignature(msg); err != nil {
		return err
	}

	if !bytes.Equal(msg.Address, c.Primary) {
		log.Printf("警告：接收到的preprepare不来自主节点")
		return nil
	}

	var preprepare *bft.Preprepare
	err := msg.Decode(&preprepare) // 直接针对msg.Msg进行解码
	if err != nil {
		return err
	}
	log.Printf("解码后得到的Preprepare：%+v", preprepare)

	if preprepare.View.Cmp(c.consensusState.View) != 0 {
		log.Printf("警告：接收到的preprepare的view不匹配")
		return nil
	}

	if preprepare.Sequence.Cmp(c.consensusState.Sequence) > 0 {
		log.Printf("接收到可能来自未来的preprepare，存储")
		c.storePreprepare(msg)
		return nil
	}

	if c.state.Cmp(pbfttypes.StatePreprepared) >= 0 {
		log.Printf("警告：当前core的状态>=StatePreprepared")
		return nil
	}
	// 接收到有效的Preprepare后，重新计时ViewChangeTimer
	//c.newViewChangeTimer()

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

	if err := c.Broadcast(&msg); err != nil {
		log.Printf("广播出现错误：%v", err)
		return err
	}

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

func (c *Core) storePreprepare(preprepareMsg *pbfttypes.Message) {
	c.pendingPrepreparesMu.Lock()
	defer c.pendingPrepreparesMu.Unlock()

	var preprepare *bft.Preprepare
	preprepareMsg.Decode(&preprepare)

	c.pendingPreprepares.Push(preprepareMsg, float32(-preprepare.Sequence.Int64()))
	log.Printf("存储preprepareMsg：%+v", preprepareMsg)
}

func (c *Core) ProcessPreprepare() {
	c.pendingPrepreparesMu.Lock()
	defer c.pendingPrepreparesMu.Unlock()

	for !(c.pendingPreprepares.Empty()) {
		m, p := c.pendingPreprepares.Pop()
		preMsg, ok := m.(*pbfttypes.Message)
		if !ok {
			log.Printf("格式不正确的preprepareMsg：%+v", m)
			continue
		}

		var pre *bft.Preprepare
		preMsg.Decode(&pre)

		if pre.Sequence.Cmp(c.consensusState.getSequence()) < 0 {
			log.Printf("pendingPreprepares中的该条prepare是旧的prepare，跳过")
			continue
		} else if pre.Sequence.Cmp(c.consensusState.getSequence()) > 0 {
			c.pendingRequests.Push(m, p)
			log.Printf("pendingPreprepares中的该条preprepare是以后需要处理的preprepare，重新放回pendingPreprepares队列")
			break
		} else {
			log.Printf("取出pendingPreprepares中的preprepare，开始准备处理……")
			c.HandlePreprepare(preMsg)
			break
		}
	}
}
