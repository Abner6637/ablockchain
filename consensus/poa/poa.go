package poa

import (
	"fmt"
)

type PoA struct {
}

func NewPoA() *PoA {
	consensus := new(PoA)

	return consensus
}

// 实现 Start 方法
func (p *PoA) Start() error {
	fmt.Println("PoA 共识已启动")
	return nil
}

// 实现 Stop 方法
func (p *PoA) Stop() error {
	fmt.Println("PoA 共识已停止")
	return nil
}
