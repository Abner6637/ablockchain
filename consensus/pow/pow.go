package pow

import (
	"fmt"
)

// Pow 实现了 Consensus 接口
type PoW struct{}

func NewPoW() *PoW {
	consensus := new(PoW)

	return consensus
}

// 实现 Start 方法
func (p *PoW) Start() error {
	fmt.Println("PoW 共识已启动")
	return nil
}

// 实现 Stop 方法
func (p *PoW) Stop() error {
	fmt.Println("PoW 共识已停止")
	return nil
}
