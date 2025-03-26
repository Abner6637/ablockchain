package pbfttypes

import (
	"log"

	"github.com/ethereum/go-ethereum/rlp"
)

const (
	MsgPreprepare uint64 = iota
	MsgPrepare
	MsgCommit
	MsgViewChange
	// msgAll
)

type Message struct {
	Code      uint64
	Msg       []byte
	Address   []byte
	Signature []byte
}

func (m *Message) EncodeMessage() ([]byte, error) {
	encodedMsg, err := rlp.EncodeToBytes(m)
	if err != nil {
		log.Fatal("RLP encoding failed:", err)
		return nil, err
	}
	return encodedMsg, nil
}

func DecodeMessage(data []byte) (*Message, error) {
	var m Message
	err := rlp.DecodeBytes(data, &m)
	if err != nil {
		log.Fatal("Failed to decode RLP data:", err)
		return nil, err
	}
	return &m, nil
}

// 解码Message.Msg
func (m *Message) Decode(val interface{}) error {
	return rlp.DecodeBytes(m.Msg, val)
}

func Encode(val interface{}) ([]byte, error) {
	return rlp.EncodeToBytes(val)
}

func (m *Message) PayloadNoSig() ([]byte, error) {
	return rlp.EncodeToBytes(&Message{
		Code:      m.Code,
		Msg:       m.Msg,
		Address:   m.Address,
		Signature: []byte{},
	})
}
