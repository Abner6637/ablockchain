package bft

import "ablockchain/core"

type ConsensusStartEvent struct {
	Block *core.Block
}

type ConsensusStopEvent struct {
	IsStop bool
}

type MessageEvent struct {
	Msg []byte
}
