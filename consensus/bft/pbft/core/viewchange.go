package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"log"
)

func (c *Core) HandleViewChange(msg *pbfttypes.Message) error {
	log.Printf("开始处理viewchange消息：%+v", msg)

	// 验证消息签名
	if err := VerifySignature(msg); err != nil {
		return err
	}

	var viewchange *bft.ViewChange
	err := msg.Decode(&viewchange) // 直接针对msg.Msg进行解码
	if err != nil {
		return err
	}
	log.Printf("解码后得到的Preprepare：%+v", viewchange)

	return nil
}
