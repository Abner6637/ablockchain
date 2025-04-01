package p2p

import (
	"ablockchain/core"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 用于传输的区块
func newTestBlock() *core.Block {
	return &core.Block{
		Header: &core.BlockHeader{
			ParentHash: []byte("0df9a8f4a2f2fc354c3c8aa5e837d4db137f20ccbf3d8336e4c95ac9d0e2943e"),
			MerkleRoot: []byte("1cdfdf5680f2a639732f6aae64a8b96c10a913b46c8fcd908c9eb95925979974"),
			Time:       uint64(time.Now().Unix()),
			Difficulty: 1,
			Nonce:      0,
			Number:     13,
		},
	}
}

func newTestTransaction() *core.Transaction {
	return &core.Transaction{
		Time:   uint64(time.Now().Unix()),
		TxHash: []byte("test_hash"),
		From:   "0xSenderAddress",
		To:     "0xReceiverAddress",
		Value:  100,
		Gas:    200,
	}
}

func TestMessageEncodingDecoding(t *testing.T) {
	msg := NewMessage(TransactionMessage, []byte("Test transaction data"))

	// 编码消息
	encodedMsg, err := msg.Encode()
	assert.Nil(t, err, "Expected no error during encoding")

	// 解码消息
	decodedMsg, err := DecodeMessage(encodedMsg)
	assert.Nil(t, err, "Expected no error during decoding")

	// 校验解码后的消息是否正确
	assert.Equal(t, msg.Type, decodedMsg.Type, "Message types should be equal")
	assert.Equal(t, msg.Data, decodedMsg.Data, "Message data should be equal")
}

// 交易
func TestProcessMessageTransaction(t *testing.T) {
	tx := newTestTransaction()
	// tx.PrintTransaction()
	data, err := tx.EncodeTx()
	if err != nil {
		return
	}
	msg := NewMessage(TransactionMessage, data)

	ProcessMessage(msg)
}

// 区块
func TestProcessMessageBlock(t *testing.T) {
	block := newTestBlock()
	// block.PrintBlock()
	data, err := block.EncodeBlock()
	if err != nil {
		return
	}
	msg := NewMessage(BlockMessage, data)

	ProcessMessage(msg)
}

// 共识
func TestProcessMessageConsensus(t *testing.T) {
	msg := NewMessage(ConsensusMessage, []byte("Test consensus"))

	ProcessMessage(msg)
}

func TestProcessMessageUnknown(t *testing.T) {
	// 创建一个未知类型的消息
	msg := NewMessage(999, []byte("Test unknown"))

	ProcessMessage(msg)
}
