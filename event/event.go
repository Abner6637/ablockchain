package event

import "ablockchain/core"

// 定义共识事件通道（传输区块）
var ConsensusStart = make(chan *core.Block, 10)

// 统一的停止通道
var stopChan = make(chan bool)

// 上链通道（传输区块）

var ConsensusFinish chan *core.Block = make(chan *core.Block, 10)

// 触发共识事件
func TriggerConsensus(block *core.Block) {
	ConsensusStart <- block
}

func TriggerStopConsensus(stop bool) {
	stopChan <- true
}

// 监听是否停止共识
func ShouldStop() bool {
	select {
	case <-stopChan:
		return true
	default:
		return false
	}
}

// 停止所有共识
func StopConsensus() {
	close(stopChan) // 关闭通道，所有监听者都会收到信号
}

// 共识完毕触发上链
func TriggerCommit(block *core.Block) {
	ConsensusFinish <- block
}
