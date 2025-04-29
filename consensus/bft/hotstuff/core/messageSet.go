package hotstuffcore

import hotstufftypes "ablockchain/consensus/bft/hotstuff/types"

// 记录每个Address对应的消息
type messageSet struct {
	messages map[string]*hotstufftypes.Message
}

func NewMessageSet() *messageSet {
	return &messageSet{
		messages: make(map[string]*hotstufftypes.Message),
	}
}
