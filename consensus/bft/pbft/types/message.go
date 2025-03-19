package pbfttypes

import (
	"log"

	"github.com/ethereum/go-ethereum/rlp"
)

const (
	MsgPreprepare uint64 = iota
	MsgPrepare
	MsgCommit
	MsgRoundChange
	// msgAll
)

type Message struct {
	Code      uint64
	Msg       []byte
	Address   []byte
	Signature []byte
}

func (m *Message) EncodeMsg() ([]byte, error) {
	encodedMsg, err := rlp.EncodeToBytes(m)
	if err != nil {
		log.Fatal("RLP encoding failed:", err)
		return nil, err
	}
	return encodedMsg, nil
}

func DecodeMsg(data []byte) (*Message, error) {
	var m Message
	err := rlp.DecodeBytes(data, &m)
	if err != nil {
		log.Fatal("Failed to decode RLP data:", err)
		return nil, err
	}
	return &m, nil
}
