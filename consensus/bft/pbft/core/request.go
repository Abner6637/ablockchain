package pbftcore

import (
	"ablockchain/consensus/bft"
	"log"
)

func (c *Core) HandleRequest(request *bft.Request) error {
	// TODO: verify Request
	log.Printf("开始处理Request：%+v", request)
	c.SendPreprepare(request)
	return nil
}
