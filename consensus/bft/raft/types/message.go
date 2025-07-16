package rafttypes

import (
	"bytes"
	"encoding/gob"
)

const (
	MsgRequestVote uint8 = iota
	MsgVoteResponse
	MsgAppendEntries
	MsgAppendResponse
)

type Message struct {
	Type        uint8
	Term        uint64
	LeaderID    []byte
	CandidateID []byte
	LogIndex    uint64
	LogTerm     uint64
	Entries     [][]byte
	CommitIndex uint64

	Success bool   // 用于回应
	Address []byte // 发送者地址
}

func (m *Message) Encode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(m)
	return buf.Bytes(), err
}

func DecodeMessage(data []byte) (*Message, error) {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	var msg Message
	err := decoder.Decode(&msg)
	return &msg, err
}
