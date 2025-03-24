package pbftcore

import pbfttypes "ablockchain/consensus/bft/pbft/types"

// 记录每个Address对应的消息
type messageSet struct {
	messages map[string]*pbfttypes.Message
}

func NewMessageSet() *messageSet {
	return &messageSet{
		messages: make(map[string]*pbfttypes.Message),
	}
}
