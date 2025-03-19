package pbftcore

import (
	pbfttypes "ablockchain/consensus/bft/pbft/types"
)

func (p *PBFT) HandlePreprepare(msg *pbfttypes.Message) error {
	preprepare, err := pbfttypes.DecodeMsg(msg.Msg)
	if err != nil {
		return nil
	}
	// TODO: 有一个结构体应该存储生成的某些消息
	if preprepare == nil {
		return nil
	}
	return nil
}

func (p *PBFT) SendPreprepare() error {
	var msg pbfttypes.Message
	msg.Code = pbfttypes.MsgPrepare

	p.Broadcast(&msg)

	return nil
}

func (p *PBFT) AcceptPrepare() {

}
