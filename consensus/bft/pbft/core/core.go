package pbftcore

import (
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"ablockchain/p2p"
)

type Core struct {
	p2pNode *p2p.Node

	consensusState *consensusState
}

func (c *Core) Start() error {

	// TODO:
	// event事件注册，监听是否有新的区块；是否有新的消息
	return nil
}

func (c *Core) Stop() error {

	// TODO
	// 关闭一些必要的进程
	return nil
}

func NewCore(p2pNode *p2p.Node) *Core {
	return &Core{
		p2pNode: p2pNode,
	}
}

func (c *Core) Broadcast(msg *pbfttypes.Message) error {
	payload, err := msg.EncodeMessage()
	if err != nil {
		return err
	}

	c.p2pNode.BroadcastMessage(string(payload))

	return nil
}
