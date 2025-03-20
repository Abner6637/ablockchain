package pbftcore

import (
	"ablockchain/consensus/bft"
)

func (c *Core) HandleRequest(request *bft.Request) error {
	// TODO: verify Request

	c.SendPreprepare(request)
	return nil
}
