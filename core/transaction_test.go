package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEncodeTransaction(t *testing.T) {
	// 创建一个测试的交易对象
	tx := Transaction{
		Time:  uint64(time.Now().Unix()),
		Hash:  []byte("test_hash"),
		From:  "0xSenderAddress",
		To:    "0xReceiverAddress",
		Value: 100,
		Gas:   200,
	}

	// 编码交易
	encodedTx, err := tx.EncodeTx()
	assert.NoError(t, err, "Failed to encode transaction")

	// 确保编码的字节流不为空
	assert.NotNil(t, encodedTx, "Encoded transaction should not be nil")

	// 验证解码后的交易数据是否匹配

	decodeTx, err := DecodeTx(encodedTx)
	assert.NoError(t, err, "Failed to decode transaction")

	// 验证收到的数据与原始数据一致
	assert.Equal(t, "0xSenderAddress", decodeTx.From, "From address should match")
	// fmt.Printf("check if equal: %s \n", decodeTx.From)
	assert.Equal(t, "0xReceiverAddress", decodeTx.To, "To address should match")
	assert.Equal(t, uint64(100), decodeTx.Value, "Value should match")
	assert.Equal(t, uint64(200), decodeTx.Gas, "Gas should match")
}
