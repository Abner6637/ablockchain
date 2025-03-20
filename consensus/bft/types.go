package bft

import (
	"math/big"
	"time"
)

type Request struct {
	Msg  []byte
	Time time.Time
}

type Preprepare struct {
	View     *big.Int
	Sequence *big.Int
	Request  Request
}

type Prepare struct {
	View     *big.Int
	Sequence *big.Int
	Digest   []byte // Request的hash
}

type Commit struct {
	View     *big.Int
	Sequence *big.Int
	Digest   []byte
}
