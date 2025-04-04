package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"ablockchain/core"
	"ablockchain/event"
	"errors"
	"fmt"
	"log"
	"math/big"
	"reflect"
	"time"
)

func (c *Core) SubcribeEvents() {
	log.Printf("订阅事件")
	c.events = []event.EventSubscription{
		{
			Name:    "ConsensusStart",
			Channel: event.Bus.Subscribe("ConsensusStart"),
		},
		{
			Name:    "ConsensusStop",
			Channel: event.Bus.Subscribe("ConsensusStop"),
		},
		{
			Name:    "MessageEvent",
			Channel: event.Bus.Subscribe("MessageEvent"),
		},
		{
			Name:    "RequestEvent",
			Channel: event.Bus.Subscribe("RequestEvent"),
		},
		{
			Name:    "FinalCommitedBlock",
			Channel: event.Bus.Subscribe("FinalCommitedBlock"),
		},
	}

	c.timeoutEvent = event.EventSubscription{
		Name:    "TimeoutEvent",
		Channel: event.Bus.Subscribe("TimeoutEvent"),
	}
}

func (c *Core) UnSubcribeEvents() {
	log.Printf("取消订阅事件")
	for _, cEvent := range c.events {
		event.Bus.Unsubscribe(cEvent.Name, cEvent.Channel)
	}
	event.Bus.Unsubscribe(c.timeoutEvent.Name, c.timeoutEvent.Channel)
}

// TODO: add a timeoutevent, and it should be out of the c.events
func (c *Core) HandleEvents() {
	go func() {
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
					c.HandleRequest(request)
				}
			case "ConsensusStop":
				log.Printf("consensus检测到ConsensusStop事件")
				if isStop, ok := eventData.(bool); ok {
					if isStop == true {
						fmt.Println("\n结束监听")
						return
					}
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

					// 待处理request中删除已经达成共识的request
					// encodedBlock, _ := block.EncodeBlock()
					// delete(c.pendingRequests, string(encodedBlock))

					// 更新共识状态，准备处理下一个区块
					c.StartNewProcess(big.NewInt(0))

					// TODO: 状态是否是在这里修改为最初状态的呢？
					c.setState(pbfttypes.StateAcceptRequest)
				}
			case "TimeoutEvent":
				log.Printf("consensus检测到TimeoutEvent事件")
				c.SendViewChange(new(big.Int).Add(c.consensusState.getView(), big.NewInt(1)))
			default:
				log.Printf("consensus未知事件类型: %s", eventNames[chosen])
			}
		}
	}()
}

func (c *Core) HandleMessage(payload []byte) error {
	msg, err := pbfttypes.DecodeMessage(payload)
	if err != nil {
		return errors.New("cannot decode message")
	}

	// TODO: different kinds of message
	switch msg.Code {
	case pbfttypes.MsgPreprepare:
		err := c.HandlePreprepare(msg)
		return err
	case pbfttypes.MsgPrepare:
		err := c.HandlePrepare(msg)
		return err
	case pbfttypes.MsgCommit:
		err := c.HandleCommit(msg)
		return err
	case pbfttypes.MsgViewChange:
		err := c.HandleViewChange(msg)
		return err
	case pbfttypes.MsgNewView:
		err := c.HandleNewView(msg)
		return err
	}

	return errors.New("invalid message")
}
