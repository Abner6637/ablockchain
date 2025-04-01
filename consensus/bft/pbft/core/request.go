package pbftcore

import (
	"ablockchain/consensus/bft"
	"ablockchain/event"
	"errors"
	"log"
)

func (c *Core) HandleRequest(request *bft.Request) error {
	log.Printf("开始处理Request：%+v", request)

	if request.GetBlockNumber() < c.consensusState.getSequence().Uint64() {
		return errors.New("old message")
	} else if request.GetBlockNumber() > c.consensusState.getSequence().Uint64() {
		c.storeRequest(request)
	} else {
		c.SendPreprepare(request)
	}

	/*
		if c.IsPrimary() {
			c.SendPreprepare(request)
		}
	*/

	return nil
}

func (c *Core) storeRequest(request *bft.Request) {
	c.pendingRequestsMu.Lock()
	defer c.pendingRequestsMu.Unlock()

	c.pendingRequests.Push(request, float32(-request.GetBlockNumber()))
}

func (c *Core) ProcessRequest() {
	c.pendingRequestsMu.Lock()
	defer c.pendingRequestsMu.Unlock()

	for !(c.pendingRequests.Empty()) {
		m, p := c.pendingRequests.Pop()
		r, ok := m.(*bft.Request)
		if !ok {
			log.Printf("格式不正确的request：%+v", m)
			continue
		}
		if r.GetBlockNumber() < c.consensusState.getSequence().Uint64() {
			continue
		} else if r.GetBlockNumber() > c.consensusState.getSequence().Uint64() {
			c.pendingRequests.Push(m, p)
			break
		} else {
			event.Bus.Publish("ConsensusStart", r.GetBlock())
		}
	}
}
