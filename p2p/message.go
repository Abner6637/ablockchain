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
	SnapshotMessage
	GetSnapshotMessage // peer节点请求状态快照信息
	GetBlocksMessage
	BlocksMessage
	HelloMessage
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
func ProcessMessage(msg *Message, peerid string) {
	switch msg.Type {
	case TransactionMessage:
		fmt.Println("处理交易消息...")
		signtx, err := core.DecodeSignTx(msg.Data)
		if err != nil {
			fmt.Printf("交易解码失败: %v\n", err)
		}
		event.Bus.Publish("TransactionMessage", signtx) //blockchain.StartMiner中处理

	case BlockMessage:
		fmt.Println("处理区块消息...")
		block, err := core.DecodeBlock(msg.Data)
		if err != nil {
			fmt.Printf("区块解码失败: %v\n", err)
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

	case GetSnapshotMessage:
		fmt.Println("GetSnapshotMessage...")
		var addr string
		err := rlp.DecodeBytes(msg.Data, &addr)
		if err != nil {
			log.Fatal("Failed to decode RLP data:", err)
		}

		event.Bus.Publish("GetSnapshotMessage", addr)

	case SnapshotMessage:
		fmt.Println("SnapshotMessage...")
		snapmsg, err := core.DecodeSnapshotMessage(msg.Data)
		if err != nil {
			fmt.Printf("解码失败: %v\n", err)
		}
		event.Bus.Publish("SnapshotMessage", snapmsg)

	case BlocksMessage:
		fmt.Println("BlocksMessage...")
		blkmsg, err := core.DecodeBlocksMessage(msg.Data)
		if err != nil {
			fmt.Printf("解码失败: %v\n", err)
		}
		event.Bus.Publish("BlocksMessage", blkmsg)

	case GetBlocksMessage:
		fmt.Println("GetBlocksMessage...")
		gblkmsg, err := core.DecodeGetBlocksMessage(msg.Data)
		if err != nil {
			fmt.Printf("解码失败: %v\n", err)
		}
		event.Bus.Publish("GetBlocksMessage", gblkmsg)

	case HelloMessage:
		log.Printf("p2p收到的消息类型为Hello消息...")
		var info EthPeerInfo
		rlp.DecodeBytes(msg.Data, &info)
		if info.PeerID == peerid { // 只处理发送者的ID和Hello消息中的ID相匹配的消息
			log.Printf("p2p接收到的Hello消息，EthAddr：%s，PeerID：%s", info.EthAddr, info.PeerID)
			log.Printf("p2p发布HelloEvent事件")
			event.Bus.Publish("HelloEvent", info)
		}
	default:
		fmt.Println("未知消息类型")
	}
}
