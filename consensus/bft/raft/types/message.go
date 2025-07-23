package rafttypes

import (
	"bytes"
	"encoding/gob"
)

const (
	MsgRequestVote = iota
	MsgVoteResponse
	MsgAppendEntries
	MsgAppendResponse
)

type Message struct {
	Code        int
	Term        uint64
	LeaderID    []byte
	CandidateID []byte
	Block       []byte // 用于AppendEntries
	VoteGranted bool
	From        []byte
	Signature   []byte
}

type AppendResp struct {
	Term    uint64
	From    string
	Success bool
}

type AppendEntries struct {
	Term     uint64
	LeaderID []byte
	Data     []byte
}

func (a *AppendEntries) Encode() []byte {
	var buf bytes.Buffer
	_ = gob.NewEncoder(&buf).Encode(a)
	return buf.Bytes()
}

func (a *AppendEntries) Decode(data []byte) error {
	return gob.NewDecoder(bytes.NewBuffer(data)).Decode(a)
}

func (r *AppendResp) Encode() []byte {
	var buf bytes.Buffer
	_ = gob.NewEncoder(&buf).Encode(r)
	return buf.Bytes()
}

func (r *AppendResp) Decode(data []byte) error {
	return gob.NewDecoder(bytes.NewBuffer(data)).Decode(r)
}
