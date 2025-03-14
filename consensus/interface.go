package consensus

// Consensus 共识协议的接口
type Consensus interface {

	// 开始挖矿/共识
	Start() error

	// 停止挖矿/共识
	Stop() error
}
