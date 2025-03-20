package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"ablockchain/event"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"
)

func (c *Core) SubcribeEvents() {
	c.events = []event.EventSubscription{
		{
			Name:    "ConsensusStart",
			Channel: event.Bus.Subscribe("ConsensusStart"),
		},
		{
			Name:    "ConsensusStop",
			Channel: event.Bus.Subscribe("ConsensusStop"),
		},
	}
}

// TODO: after events finished
func (c *Core) HandleEvents() {
	go func() {
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
				if startEvent, ok := eventData.(bft.ConsensusStartEvent); ok { // 类型断言确保事件数据类型的正确性
					request := &bft.Request{
						Msg:  startEvent.Block.Hash,
						Time: time.Now(),
					}
					c.HandleRequest(request)
				}
			case "ConsensusStop":
				if stopEvent, ok := eventData.(bft.ConsensusStopEvent); ok {
					if stopEvent.IsStop == true {
						fmt.Println("\n结束监听")
						return
					}
				}
			default:
				log.Printf("未知事件类型: %s", eventNames[chosen])
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
	}

	return errors.New("invalid message")
}
