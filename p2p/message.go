package p2p

import (
	"ablockchain/core"
	"ablockchain/event"
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"
)

const (
	TransactionMessage = iota
	BlockMessage
	ConsensusMessage
)

// 消息结构体
type Message struct {
	Type uint64
	Data []byte
}

// 封装消息
func NewMessage(msgType uint64, data []byte) *Message {
	return &Message{
		Type: msgType,
		Data: data,
	}
}

// 编码消息
func (msg *Message) Encode() ([]byte, error) {
	encodedMessage, err := rlp.EncodeToBytes(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to encode message: %v", err)
	}
	return encodedMessage, nil
}

// 解码消息
func DecodeMessage(data []byte) (*Message, error) {
	var msg Message
	err := rlp.DecodeBytes(data, &msg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode message: %v", err)
	}
	return &msg, nil
}

// 分类处理消息
func ProcessMessage(msg *Message) {
	switch msg.Type {
	case TransactionMessage:
		fmt.Println("处理交易消息...")
		tx, err := core.DecodeTx(msg.Data)
		if err != nil {
			fmt.Errorf("交易解码失败")
		}
		tx.PrintTransaction()
		event.Bus.Publish("TransactionMessage", tx)

	case BlockMessage:
		fmt.Println("处理区块消息...")
		block, err := core.DecodeBlock(msg.Data)
		if err != nil {
			fmt.Errorf("区块解码失败")
		}
		block.PrintBlock()
		event.Bus.Publish("BlockMessage", block)

	case ConsensusMessage:
		fmt.Println("处理共识消息...")
	default:
		fmt.Println("未知消息类型")
	}
}
