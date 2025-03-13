package bft

import "math/big"

type Proposal struct {
	Number *big.Int
	hash   []byte
}

type Request struct {
	Proposal Proposal
}

type View struct {
	Round    *big.Int
	Sequence *big.Int
}

type Preprepare struct {
	View     *View
	Proposal Proposal
}

type Subject struct {
	View   *View
	Digest []byte
}
