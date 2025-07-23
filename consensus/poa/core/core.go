package poacore

import (
	poatypes "ablockchain/consensus/poa/types"
	"ablockchain/core"
	"ablockchain/crypto"
	"ablockchain/event"
	"ablockchain/p2p"
	"bytes"
	"crypto/ecdsa"
	"encoding/gob"
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

func init() {
	gob.Register(&poatypes.ValidatorMessage{})
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
		{Name: "ConsensusStart", Channel: event.Bus.Subscribe("ConsensusStart")},
		{Name: "BlockMessage", Channel: event.Bus.Subscribe("BlockMessage")},
		{Name: "ConsensusStop", Channel: event.Bus.Subscribe("ConsensusStop")},
		{Name: "MessageEvent", Channel: event.Bus.Subscribe("MessageEvent")},
	}
}

// HandleEvents 监听事件
func (c *Core) HandleEvents() {
	go func() {
		cases := make([]reflect.SelectCase, len(c.events))
		eventNames := make([]string, len(c.events))
		for i, sub := range c.events {
			cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(sub.Channel)}
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

			case "MessageEvent":
				if raw, ok := eventData.([]byte); ok {
					c.HandleConsensusMessage(raw)
				}

			default:
				log.Printf("[PoA] 未知事件: %s", eventNames[chosen])
			}
		}
	}()
}

// HandleConsensusMessage 处理验证者同步消息
func (c *Core) HandleConsensusMessage(payload []byte) {
	msg, err := poatypes.DecodeValidatorMessage(payload)
	if err != nil {
		log.Printf("[PoA] 解码验证者消息失败: %v", err)
		return
	}

	addrBytes, err := hex.DecodeString(msg.Address) // 解码 hex 字符串为 []byte
	if err != nil {
		log.Printf("[PoA] 地址解码失败: %v", err)
		return
	}

	switch msg.Code {
	case poatypes.MsgAddValidator:
		log.Printf("[PoA] 收到新增验证者广播: 0x%x", addrBytes)
		c.addValidatorInternal(addrBytes)
	case poatypes.MsgRemoveValidator:
		log.Printf("[PoA] 收到删除验证者广播: 0x%x", addrBytes)
		c.removeValidatorInternal(addrBytes)
	}
}

func (c *Core) broadcastValidatorMessage(code uint8, address []byte) {
	msg := &poatypes.ValidatorMessage{
		Code:    code,
		Address: hex.EncodeToString(address),
	}
	encoded, err := msg.Encode()
	if err != nil {
		log.Printf("[PoA] 编码验证者消息失败: %v", err)
		return
	}
	p2pMsg := &p2p.Message{Type: p2p.ConsensusMessage, Data: encoded}
	finalBytes, err := p2pMsg.Encode()
	if err != nil {
		log.Printf("[PoA] 编码 P2P 消息失败: %v", err)
		return
	}
	c.p2pNode.BroadcastMessage(string(finalBytes))
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
	block.Header.Extra = c.address

	encodedBlock, err := block.EncodeBlock()
	if err != nil {
		log.Printf("[PoA] 区块编码失败: %v", err)
		return
	}

	p2pMsg := &p2p.Message{Type: p2p.BlockMessage, Data: encodedBlock}
	encodedP2PMsg, err := p2pMsg.Encode()
	if err != nil {
		log.Printf("[PoA] 消息编码失败: %v", err)
		return
	}
	c.p2pNode.BroadcastMessage(string(encodedP2PMsg))
	log.Printf("[PoA] 广播已完成")

	c.OnReceiveBlock(block)
}

func (c *Core) OnReceiveBlock(block *core.Block) {
	if c.ValidateBlock(block) {
		log.Printf("[PoA] 区块验证通过")
		event.Bus.Publish("ConsensusFinish", block)
	} else {
		log.Printf("[PoA] 区块验证失败")
	}
}

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

func (c *Core) AddValidator(newHex string) {
	newHex = strings.TrimPrefix(newHex, "0x")
	addrBytes, err := hex.DecodeString(newHex)
	if err != nil {
		log.Printf("[PoA] 新增验证者解析失败: %v", err)
		return
	}
	c.addValidatorInternal(addrBytes)
	c.broadcastValidatorMessage(poatypes.MsgAddValidator, addrBytes)
}

func (c *Core) RemoveValidator(addrHex string) {
	addrHex = strings.TrimPrefix(addrHex, "0x")
	addrBytes, err := hex.DecodeString(addrHex)
	if err != nil {
		log.Printf("[PoA] 删除验证者解析失败: %v", err)
		return
	}
	c.removeValidatorInternal(addrBytes)
	c.broadcastValidatorMessage(poatypes.MsgRemoveValidator, addrBytes)
}

func (c *Core) addValidatorInternal(addr []byte) {
	for _, v := range c.ValSet {
		if bytes.Equal(v, addr) {
			log.Printf("[PoA] 验证者已存在: 0x%x", addr)
			return
		}
	}
	c.ValSet = append(c.ValSet, addr)
	sort.Slice(c.ValSet, func(i, j int) bool {
		return bytes.Compare(c.ValSet[i], c.ValSet[j]) < 0
	})
	log.Printf("[PoA] 新增验证者: 0x%x, 新 ValSet: %+v", addr, c.ValSet)
}

func (c *Core) removeValidatorInternal(addr []byte) {
	for i, v := range c.ValSet {
		if bytes.Equal(v, addr) {
			c.ValSet = append(c.ValSet[:i], c.ValSet[i+1:]...)
			sort.Slice(c.ValSet, func(i, j int) bool {
				return bytes.Compare(c.ValSet[i], c.ValSet[j]) < 0
			})
			log.Printf("[PoA] 已删除验证者: 0x%x, 新 ValSet: %+v", addr, c.ValSet)
			return
		}
	}
	log.Printf("[PoA] 验证者不存在: 0x%x", addr)
}
