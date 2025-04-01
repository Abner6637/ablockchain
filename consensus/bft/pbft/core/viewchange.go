package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"log"
	"math/big"
)

func (c *Core) HandleViewChange(msg *pbfttypes.Message) error {
	log.Printf("开始处理viewchange消息：%+v", msg)

	// 验证消息签名
	if err := VerifySignature(msg); err != nil {
		return err
	}

	var viewchange *bft.ViewChange
	err := msg.Decode(&viewchange)
	if err != nil {
		return err
	}
	log.Printf("解码后得到的Preprepare：%+v", viewchange)

	newView := viewchange.View
	c.addViewChange(newView, msg)

	// 当自己为新视图下的主节点时，发送NewView消息，同时也要更新自己的状态
	if len(c.ViewChanges[newView.Uint64()].messages) >= 2*c.ByzantineSize+1 && string(c.address) == c.PrimaryFromView(viewchange.View) && newView.Cmp(c.consensusState.View) > 0 {
		c.SendNewView(newView)
		c.StartNewProcess(newView)
	}

	return nil
}

func (c *Core) SendViewChange(view *big.Int) error {
	var msg pbfttypes.Message
	msg.Code = pbfttypes.MsgViewChange

	viewchange := c.consensusState.getViewChange(view)

	encodedviewchange, err := pbfttypes.Encode(viewchange)
	if err != nil {
		return err
	}

	// 打印使用
	log.Printf("生成ViewChange：%+v", viewchange)

	msg.Msg = encodedviewchange
	msg.Address = c.address
	msg.Signature, err = c.SignMessage(&msg)
	if err != nil {
		return err
	}

	// 发给别人ViewChange消息时，自己也保存一份自己发送的
	c.addViewChange(view, &msg)

	log.Printf("广播ViewChange消息：%+v", msg)
	c.Broadcast(&msg)

	return nil
}
