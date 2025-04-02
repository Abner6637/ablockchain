package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"ablockchain/core"
	"ablockchain/event"
	"errors"
	"log"
	"math/big"
	"reflect"
	"sync"
	"testing"
	"time"

	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
)

func TestProcessRequest(t *testing.T) {
	c := &Core{
		state:             pbfttypes.StateAcceptRequest,
		consensusState:    NewConsensusState(big.NewInt(0), big.NewInt(0), nil),
		pendingRequests:   prque.New(),
		pendingRequestsMu: new(sync.Mutex),
		curCommitedBlock:  &core.Block{},
	}

	var requests []*bft.Request

	for i := 0; i < 3; i++ {
		block := newTestBlock(big.NewInt(int64(i)))
		encodedBlock, _ := block.EncodeBlock()
		requests = append(requests, &bft.Request{
			Msg:  encodedBlock,
			Time: uint64(time.Now().Unix()),
		})
	}

	c.storeRequest(requests[1])
	c.storeRequest(requests[0])
	c.storeRequest(requests[2])

	/* 测试优先级队列是否正常
	for !(c.pendingRequests.Empty()) {
		r, p := c.pendingRequests.Pop()
		log.Printf("request: %+v, priority: %v", r, p)
	}
	c.storeRequest(requests[1])
	c.storeRequest(requests[0])
	c.storeRequest(requests[2])
	*/

	c.SubcribeEvents()
	defer c.UnSubcribeEvents()

	go c.testHandleEventsForRequest()

	c.ProcessRequest()
	time.Sleep(3 * time.Second)

}

func (c *Core) handleRequestForTest(request *bft.Request) error {
	log.Printf("开始处理Request：%+v", request)

	if request.GetBlockNumber().Cmp(c.consensusState.getSequence()) < 0 {
		return errors.New("old message")
	} else if request.GetBlockNumber().Cmp(c.consensusState.getSequence()) > 0 {
		c.storeRequest(request)
	} else {
		log.Printf("处理正确的Request……")

		// 共识过程省略；共识完毕后，发送事件
		// 对于改事件的处理会重新调用c.ProcessRequest()  （因为没有视图切换）
		event.Bus.Publish("FinalCommitedBlock", request.GetBlock())
	}

	/*
		if c.IsPrimary() {
			c.SendPreprepare(request)
		}
	*/

	return nil
}

func (c *Core) testHandleEventsForRequest() {
	log.Printf("开始处理事件")
	// 准备反射监听参数
	cases := make([]reflect.SelectCase, len(c.events))
	eventNames := make([]string, len(c.events)) // 保存对应的名称

	for i, sub := range c.events {
		cases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(sub.Channel),
		}
		eventNames[i] = sub.Name // 建立索引与名称的映射
	}

	cases = append(cases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(c.timeoutEvent.Channel),
	})
	eventNames = append(eventNames, c.timeoutEvent.Name)

	for {
		// 阻塞监听事件
		chosen, value, ok := reflect.Select(cases)

		// 处理通道关闭
		if !ok {
			log.Printf("事件通道关闭: %s", eventNames[chosen])
			return
		}

		// 获取事件数据
		eventData := value.Interface()

		// 根据事件名称路由处理
		switch eventNames[chosen] {
		case "ConsensusStart":
			log.Printf("consensus检测到ConsensusStart事件")
			if block, ok := eventData.(*core.Block); ok { // 类型断言确保事件数据类型的正确性
				log.Printf("接收到的block：%+v", block)
				log.Printf("接收到的block的header：%+v", block.Header)
				encodedBlock, _ := block.EncodeBlock()
				request := &bft.Request{
					Msg:  encodedBlock,
					Time: uint64(time.Now().Unix()),
				}
				c.handleRequestForTest(request)
			}
		case "MessageEvent":
			log.Printf("consensus检测到MessageEvent事件")
			if msg, ok := eventData.([]byte); ok {
				err := c.HandleMessage(msg)
				if err != nil {
					log.Println("failed to handle the message")
				}
			}
		case "FinalCommitedBlock":
			log.Printf("consensus检测到FinalCommitedBlock事件")
			if block, ok := eventData.(*core.Block); ok {
				// 最新达成共识的区块
				c.curCommitedBlock = block
				log.Printf("最新达成共识的区块的header：%+v", c.curCommitedBlock.Header)

				c.consensusState = NewConsensusState(c.consensusState.getView(), new(big.Int).Add(c.curCommitedBlock.Header.Number, big.NewInt(1)), nil)
				c.ProcessRequest()

				c.setState(pbfttypes.StateAcceptRequest)
			}
		default:
			log.Printf("consensus未知事件类型: %s", eventNames[chosen])
		}
	}
}
