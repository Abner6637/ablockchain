package bft

import (
	"ablockchain/crypto"
	"bytes"
	"encoding/binary"
	"math/big"
	"time"
)

type Request struct {
	Msg  []byte
	Time time.Time
}

func (r *Request) HashReqeust() []byte {
	var buf bytes.Buffer
	buf.Write(r.Msg)
	binary.Write(&buf, binary.BigEndian, uint32(r.Time.Unix()))

	return crypto.GlobalHashAlgorithm.Hash(buf.Bytes())
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

func (p *Prepare) HashPrepare() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, p.View)
	binary.Write(&buf, binary.BigEndian, p.Sequence)
	buf.Write(p.Digest)

	return crypto.GlobalHashAlgorithm.Hash(buf.Bytes())
}

type Commit struct {
	View     *big.Int
	Sequence *big.Int
	Digest   []byte
}

func (c *Commit) HashCommit() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, c.View)
	binary.Write(&buf, binary.BigEndian, c.Sequence)
	buf.Write(c.Digest)

	return crypto.GlobalHashAlgorithm.Hash(buf.Bytes())
}
