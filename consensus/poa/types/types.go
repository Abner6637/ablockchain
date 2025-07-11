package poatypes

/* import (
	"bytes"
	"encoding/gob"
)

// 消息类型常量
const (
	MsgPreprepare uint8 = iota
	MsgPrepare
	MsgCommit
)

// PoA 消息体
type Message struct {
	Code      uint8  // 消息类型
	Msg       []byte // 消息体（区块hash 或 序列化后的区块）
	Address   []byte // 发送者地址
	Signature []byte // 签名
}

// EncodeMessage 序列化
func (m *Message) EncodeMessage() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(m)
	return buf.Bytes(), err
}

// DecodeMessage 反序列化
func DecodeMessage(data []byte) (*Message, error) {
	var m Message
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&m)
	return &m, err
}

// 生成签名前的载荷（去掉 Signature）
func (m *Message) PayloadNoSig() ([]byte, error) {
	tmp := &Message{
		Code:    m.Code,
		Msg:     m.Msg,
		Address: m.Address,
		// Signature 不要
	}
	return tmp.EncodeMessage()
}
*/
