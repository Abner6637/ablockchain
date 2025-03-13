package core

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
)

func TestEncodeTransaction(t *testing.T) {
	// 创建一个测试的交易对象
	tx := Transaction{
		Time:  time.Now(),
		Hash:  []byte("test_hash"),
		From:  "0xSenderAddress",
		To:    "0xReceiverAddress",
		Value: 100,
		Gas:   200,
	}

	// 编码交易
	encodedTx, err := encodeTransaction(tx)
	assert.NoError(t, err, "Failed to encode transaction")

	// 确保编码的字节流不为空
	assert.NotNil(t, encodedTx, "Encoded transaction should not be nil")

	// 验证解码后的交易数据是否匹配
	var txr Transaction
	err = rlp.DecodeBytes(encodedTx, &txr)
	assert.NoError(t, err, "Failed to decode transaction")

	// 验证收到的数据与原始数据一致
	assert.Equal(t, "0xSenderAddress", txr.From, "From address should match")
	assert.Equal(t, "0xReceiverAddress", txr.To, "To address should match")
	assert.Equal(t, uint64(100), txr.Value, "Value should match")
	assert.Equal(t, uint64(200), txr.Gas, "Gas should match")
}
