package p2p

import (
	"ablockchain/core"
	"ablockchain/event"
	"fmt"
	"log"

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
		signtx, err := core.DecodeSignTx(msg.Data)
		if err != nil {
			fmt.Errorf("交易解码失败")
		}
		event.Bus.Publish("TransactionMessage", signtx)

	case BlockMessage:
		fmt.Println("处理区块消息...")
		block, err := core.DecodeBlock(msg.Data)
		if err != nil {
			fmt.Errorf("区块解码失败")
		}
		// block.PrintBlock()
		event.Bus.Publish("BlockMessage", block)

	case ConsensusMessage:
		log.Printf("p2p收到的消息类型为共识消息...")
		/* 测试TestMessageHandler时使用
		var mmsg string
		rlp.DecodeBytes(msg.Data, &mmsg)
		log.Printf("接收到的消息，类型：%d，数据为编码后的：%s", msg.Type, mmsg)
		*/
		log.Printf("p2p发布MessageEvent事件")
		event.Bus.Publish("MessageEvent", msg.Data)
	default:
		fmt.Println("未知消息类型")
	}
}
