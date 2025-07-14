package poacore

import (
	"ablockchain/core"
	"ablockchain/crypto"
	"ablockchain/event"
	"ablockchain/p2p"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"log"
	"reflect"
	"sort"
	"strings"
)

// Core 定义 PoA 共识核心
type Core struct {
	p2pNode    *p2p.Node
	blockchain *core.Blockchain
	privateKey *ecdsa.PrivateKey
	address    []byte
	ValSet     [][]byte // 存二进制地址（[]byte）
	events     []event.EventSubscription
	running    bool
}

// NewCore 创建 PoA 核心
func NewCore(p2pNode *p2p.Node, bc *core.Blockchain, valSet []string) *Core {
	address := crypto.PubkeyToAddress(p2pNode.PrivateKey.PublicKey).Bytes()

	log.Printf("[PoA] 我自己地址: 0x%x", address)

	var realSet [][]byte
	if len(valSet) == 0 {
		realSet = append(realSet, address)
	} else {
		for _, v := range valSet {
			v = strings.TrimPrefix(v, "0x")
			addrBytes, err := hex.DecodeString(v)
			if err != nil {
				log.Fatalf("[PoA] ValSet 解析失败: %v", err)
			}
			realSet = append(realSet, addrBytes)
		}
	}

	// 排序地址切片，保证一致顺序
	sort.Slice(realSet, func(i, j int) bool {
		return bytes.Compare(realSet[i], realSet[j]) < 0
	})

	log.Printf("[PoA] ValSet: %+v", realSet)

	return &Core{
		p2pNode:    p2pNode,
		blockchain: bc,
		privateKey: p2pNode.PrivateKey,
		address:    address,
		ValSet:     realSet,
	}
}

// Start 启动共识
func (c *Core) Start() error {
	if c.running {
		log.Println("[PoA] 已在运行")
		return nil
	}

	log.Printf("[PoA] 启动，共识地址: 0x%x", c.address)
	c.running = true
	c.SubscribeEvents()
	c.HandleEvents()
	return nil
}

// Stop 停止共识
func (c *Core) Stop() error {
	if !c.running {
		return nil
	}
	c.running = false
	event.Bus.Publish("ConsensusStop", true)
	log.Println("[PoA] 已停止")
	return nil
}

// Address 返回节点地址
func (c *Core) Address() []byte {
	return c.address
}

// SubscribeEvents 订阅事件
func (c *Core) SubscribeEvents() {
	c.events = []event.EventSubscription{
		{
			Name:    "ConsensusStart",
			Channel: event.Bus.Subscribe("ConsensusStart"),
		},
		{
			Name:    "BlockMessage",
			Channel: event.Bus.Subscribe("BlockMessage"),
		},
		{
			Name:    "ConsensusStop",
			Channel: event.Bus.Subscribe("ConsensusStop"),
		},
	}
}

// HandleEvents 监听事件
func (c *Core) HandleEvents() {
	go func() {
		cases := make([]reflect.SelectCase, len(c.events))
		eventNames := make([]string, len(c.events))
		for i, sub := range c.events {
			cases[i] = reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(sub.Channel),
			}
			eventNames[i] = sub.Name
		}

		for {
			chosen, value, ok := reflect.Select(cases)
			if !ok {
				log.Printf("[PoA] 事件通道关闭: %s", eventNames[chosen])
				return
			}
			eventData := value.Interface()

			switch eventNames[chosen] {
			case "ConsensusStart":
				block, ok := eventData.(*core.Block)
				if !ok {
					log.Fatal("[PoA] ConsensusStart 转换失败")
				}
				go c.Propose(block)

			case "BlockMessage":
				block, ok := eventData.(*core.Block)
				if !ok {
					log.Fatal("[PoA] BlockMessage 转换失败")
				}
				c.OnReceiveBlock(block)

			case "ConsensusStop":
				log.Println("[PoA] 收到停止信号")
				return

			default:
				log.Printf("[PoA] 未知事件: %s", eventNames[chosen])
			}
		}
	}()
}

// Propose 轮值验证并广播区块
func (c *Core) Propose(block *core.Block) {
	height := c.blockchain.CurrentBlock().Header.Number.Uint64() + 1
	validatorIndex := height % uint64(len(c.ValSet))
	expected := c.ValSet[validatorIndex]

	if !bytes.Equal(expected, c.address) {
		log.Printf("[PoA] 轮到验证者: 0x%x, 我是: 0x%x，不是我，跳过", expected, c.address)
		return
	}

	log.Printf("[PoA] 轮到我出块 #%d", height)

	// Extra 写裸地址
	block.Header.Extra = c.address

	encodedBlock, err := block.EncodeBlock()
	if err != nil {
		log.Printf("[PoA] 区块编码失败: %v", err)
		return
	}

	p2pMsg := &p2p.Message{
		Type: p2p.BlockMessage,
		Data: encodedBlock,
	}
	encodedP2PMsg, err := p2pMsg.Encode()
	if err != nil {
		log.Printf("[PoA] 消息编码失败: %v", err)
		return
	}

	c.p2pNode.BroadcastMessage(string(encodedP2PMsg))
	log.Printf("[PoA] 广播已完成")

	// 本节点也验证
	c.OnReceiveBlock(block)
}

// OnReceiveBlock 收到区块验证并执行
func (c *Core) OnReceiveBlock(block *core.Block) {
	if c.ValidateBlock(block) {
		log.Printf("[PoA] 区块验证通过")
		event.Bus.Publish("ConsensusFinish", block)
	} else {
		log.Printf("[PoA] 区块验证失败")
	}
}

// ValidateBlock 验证轮值合法性
func (c *Core) ValidateBlock(block *core.Block) bool {
	height := block.Header.Number.Uint64()
	validatorIndex := height % uint64(len(c.ValSet))

	expected := c.ValSet[validatorIndex]
	actual := block.Header.Extra

	if !bytes.Equal(expected, actual) {
		log.Printf("[PoA] 验证失败，预期: 0x%x, 实际: 0x%x", expected, actual)
		return false
	}
	return true
}

// AddValidator 新增验证者（可选）
func (c *Core) AddValidator(newHex string) {
	newHex = strings.TrimPrefix(newHex, "0x")
	addrBytes, err := hex.DecodeString(newHex)
	if err != nil {
		log.Printf("[PoA] 新增验证者解析失败: %v", err)
		return
	}

	for _, v := range c.ValSet {
		if bytes.Equal(v, addrBytes) {
			log.Printf("[PoA] 验证者已存在: 0x%x", addrBytes)
			return
		}
	}
	c.ValSet = append(c.ValSet, addrBytes)
	sort.Slice(c.ValSet, func(i, j int) bool {
		return bytes.Compare(c.ValSet[i], c.ValSet[j]) < 0
	})
	log.Printf("[PoA] 新增验证者: 0x%x, 新 ValSet: %+v", addrBytes, c.ValSet)
}
