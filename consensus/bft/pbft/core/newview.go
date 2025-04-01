package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"bytes"
	"log"
	"math/big"
)

func (c *Core) HandleNewView(msg *pbfttypes.Message) error {
	log.Printf("开始处理newview消息：%+v", msg)

	// 验证消息签名
	if err := VerifySignature(msg); err != nil {
		return err
	}

	var newview *bft.NewView
	err := msg.Decode(&newview) // 直接针对msg.Msg进行解码
	if err != nil {
		return err
	}
	log.Printf("解码后得到的NewView：%+v", newview)

	if !bytes.Equal(msg.Address, []byte(c.PrimaryFromView(newview.View))) {
		log.Printf("警告：接收到的newview不来自新的主节点")
		return nil
	}

	c.StartNewProcess(newview.View)

	// 当有正在进行共识的区块时，向其他节点发送Preprepare消息，重新对该区块进行共识
	if c.consensusState.getPreprepare() != nil {
		c.setState(pbfttypes.StatePreprepared)
		c.SendPrepare()
	}

	return nil
}

func (c *Core) SendNewView(view *big.Int) error {
	var msg pbfttypes.Message
	msg.Code = pbfttypes.MsgNewView

	newview := c.consensusState.getNewView(view)

	encodednewview, err := pbfttypes.Encode(newview)
	if err != nil {
		return err
	}

	// 打印使用
	log.Printf("生成newview：%+v", newview)

	msg.Msg = encodednewview
	msg.Address = c.address
	msg.Signature, err = c.SignMessage(&msg)
	if err != nil {
		return err
	}

	log.Printf("广播NewView消息：%+v", msg)
	c.Broadcast(&msg)

	return nil
}
