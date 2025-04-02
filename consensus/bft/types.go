package bft

import (
	"ablockchain/core"
	"ablockchain/crypto"
	"bytes"
	"encoding/binary"
	"math/big"
)

type Request struct {
	Msg  []byte // 这里存储编码后的区块比存储区块的哈希更好，选择存储编码后的区块
	Time uint64
}

func (r *Request) GetBlockNumber() *big.Int {
	block, _ := core.DecodeBlock(r.Msg)
	return block.Header.Number
}

func (r *Request) GetBlock() *core.Block {
	block, _ := core.DecodeBlock(r.Msg)
	return block
}

func (r *Request) HashReqeust() []byte {
	var buf bytes.Buffer
	buf.Write(r.Msg)
	binary.Write(&buf, binary.BigEndian, r.Time)

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

type ViewChange struct {
	View     *big.Int // 新视图编号
	Sequence *big.Int
}

type NewView struct {
	View     *big.Int // 新视图编号
	Sequence *big.Int // 新序列号
}
