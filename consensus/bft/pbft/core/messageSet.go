package pbftcore

import pbfttypes "ablockchain/consensus/bft/pbft/types"

type messageSet struct {
	messages map[string]*pbfttypes.Message
}
