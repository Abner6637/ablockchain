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
}

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
				if block, ok := eventData.(*core.Block); ok { // 类型断言确保事件数据类型的正确性
					request := &bft.Request{
						Msg:  block.Hash,
						Time: time.Now(),
					}
					// TODO，目前Request是直接由自己生成的（假设是主节点的话）
					// 其次，目前还没有做主节点区分（怎么区分？）
					c.HandleRequest(request)
				}
			case "ConsensusStop":
				if isStop, ok := eventData.(bool); ok {
					if isStop == true {
						fmt.Println("\n结束监听")
						return
					}
				}
			case "MessageEvent":
				if msg, ok := eventData.([]byte); ok {
					err := c.HandleMessage(msg)
					if err != nil {
						log.Println("failed to handle the message")
					}
				}
			case "FinalCommitedBlock":
				if block, ok := eventData.(*core.Block); ok {
					c.curCommitedBlock = block
					c.StartNewProcess(big.NewInt(0))
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
	case pbfttypes.MsgPrepare:
		err := c.HandlePrepare(msg)
		return err
	case pbfttypes.MsgCommit:
		err := c.HandleCommit(msg)
		return err
	}

	return errors.New("invalid message")
}
