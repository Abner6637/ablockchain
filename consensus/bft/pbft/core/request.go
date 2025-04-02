package pbftcore

import (
	"ablockchain/consensus/bft"
	"ablockchain/event"
	"errors"
	"log"
)

func (c *Core) HandleRequest(request *bft.Request) error {
	log.Printf("开始处理Request：%+v", request)

	if request.GetBlockNumber().Cmp(c.consensusState.getSequence()) < 0 {
		return errors.New("old message")
	} else if request.GetBlockNumber().Cmp(c.consensusState.getSequence()) > 0 {
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

	c.pendingRequests.Push(request, float32(-request.GetBlockNumber().Int64()))
	log.Printf("存储request：%+v", request)
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
		if r.GetBlockNumber().Cmp(c.consensusState.getSequence()) < 0 {
			log.Printf("pendingRequest中的该条request是旧的request，跳过")
			continue
		} else if r.GetBlockNumber().Cmp(c.consensusState.getSequence()) > 0 {
			c.pendingRequests.Push(m, p)
			log.Printf("pendingRequest中的该条request是以后需要处理的request，重新放回pendingRequest队列")
			break
		} else {
			log.Printf("取出pendingRequest中的request，开始准备处理……")
			event.Bus.Publish("ConsensusStart", r.GetBlock())
			break
		}
	}
}
