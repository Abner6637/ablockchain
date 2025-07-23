package poatypes

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"
)

// 消息类型枚举
const (
	MsgAddValidator    = uint8(1)
	MsgRemoveValidator = uint8(2)
)

// 验证者变更消息结构
type ValidatorMessage struct {
	Code    uint8  // 消息类型：1 表示添加，2 表示移除
	Address string // 验证者地址（16进制字符串）
}

// 编码消息
func (msg *ValidatorMessage) Encode() ([]byte, error) {
	var buf bytes.Buffer
	err := rlp.Encode(&buf, msg)
	return buf.Bytes(), err
}

// 解码消息
func DecodeValidatorMessage(data []byte) (*ValidatorMessage, error) {
	var msg ValidatorMessage
	err := rlp.DecodeBytes(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// 打印日志使用
func (msg *ValidatorMessage) String() string {
	return fmt.Sprintf("ValidatorMessage{Code=%d, Address=%s}", msg.Code, msg.Address)
}

// 工具方法：地址格式验证（可选）
func IsValidHexAddress(address string) bool {
	_, err := hex.DecodeString(address[2:]) // 0x开头
	return err == nil
}
