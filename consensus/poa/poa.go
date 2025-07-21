//package poa

//import (
//	"fmt"
//)

//type PoA struct {
//}

//func NewPoA() *PoA {
//	consensus := new(PoA)

//	return consensus
//}

// 实现 Start 方法
//func (p *PoA) Start() error {
//	fmt.Println("PoA 共识已启动")
//	return nil
//}

// 实现 Stop 方法
//
//	func (p *PoA) Stop() error {
//		fmt.Println("PoA 共识已停止")
//		return nil
//	}

package poacore

import (
	"ablockchain/cli"
	"ablockchain/core"
	"ablockchain/crypto"
	"ablockchain/p2p"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

const sealInterval = 30 * time.Second

type Core struct {
	p2pNode         *p2p.Node
	blockchain      *core.Blockchain
	privateKey      *ecdsa.PrivateKey
	address         []byte
	ValSet          []string
	wg              *sync.WaitGroup
	stopChan        chan struct{}
	lastValidator   string        // 记录上次的验证者地址
	currentRound    uint64        // 记录当前轮次
	lastSkipLogTime time.Time     // 记录上次打印"跳过打包"日志的时间
	skipLogInterval time.Duration // 跳过日志的最小时间间隔
	lastSkipReason  string        // 记录上次跳过打包的原因
}

// 创建共识核心
func NewCore(cfg *cli.Config, p2pNode *p2p.Node, bc *core.Blockchain) *Core {
	address := crypto.PubkeyToAddress(p2pNode.PrivateKey.PublicKey).Bytes()

	normalizeAddr := func(addr string) string {
		return strings.ToLower(strings.TrimPrefix(addr, "0x"))
	}

	var valSet []string
	if len(cfg.ValSet) > 0 {
		for _, addr := range cfg.ValSet {
			valSet = append(valSet, normalizeAddr(addr))
		}
	} else {
		valSet = append(valSet, hex.EncodeToString(address))
	}

	log.Printf("[PoA] 配置验证者集合: %v", valSet)
	log.Printf("[PoA] 当前节点地址: %s", hex.EncodeToString(address))

	return &Core{
		p2pNode:         p2pNode,
		blockchain:      bc,
		privateKey:      p2pNode.PrivateKey,
		address:         address,
		ValSet:          valSet,
		wg:              &sync.WaitGroup{},
		stopChan:        make(chan struct{}),
		lastValidator:   "",              // 初始化为空
		currentRound:    0,               // 初始轮次
		lastSkipLogTime: time.Now(),      // 初始化日志时间
		skipLogInterval: 2 * time.Minute, // 设置2分钟间隔
		lastSkipReason:  "",              // 初始化跳过原因
	}
}

// 启动共识
func (c *Core) Start() error {
	log.Printf("[PoA] 启动，共识地址: 0x%x", c.Address())

	// 注册接收消息处理器
	c.p2pNode.OnMessage = c.HandleP2PMessage

	c.wg.Add(1)
	go c.loopSeal()
	return nil
}

// 停止共识
func (c *Core) Stop() error {
	close(c.stopChan)
	c.wg.Wait()
	log.Println("[PoA] 停止")
	return nil
}

// 返回地址
func (c *Core) Address() []byte {
	return c.address
}

// 出块轮询定时器
func (c *Core) loopSeal() {
	defer c.wg.Done()

	ticker := time.NewTicker(sealInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.trySeal()
		case <-c.stopChan:
			return
		}
	}
}

// 出块尝试
func (c *Core) trySeal() {
	currentBlock := c.blockchain.CurrentBlock()
	if currentBlock == nil {
		c.logSkipped("无法获取当前区块", true)
		return
	}
	height := currentBlock.Header.Number.Uint64()
	nextHeight := height + 1

	// 计算当前轮次和验证者
	round := nextHeight / uint64(len(c.ValSet))
	validatorIndex := nextHeight % uint64(len(c.ValSet))
	expected := c.ValSet[validatorIndex]
	currentAddr := hex.EncodeToString(c.address)

	// 只在轮次变化或验证者变化时打印日志
	if round != c.currentRound || expected != c.lastValidator {
		log.Printf("[PoA] 轮次 #%d - 当前轮到验证者: 0x%s", round, expected)
		c.currentRound = round
		c.lastValidator = expected
	}

	// 检查是否轮到本节点
	if expected != currentAddr {
		return
	}

	// 检查交易池
	if c.blockchain.TxPool.PendingSize() == 0 {
		c.logSkipped("没有待处理交易", false)
		return
	}

	log.Printf("[PoA] 当前轮到我 打包 块 #%d", nextHeight)

	minerAddr := c.blockchain.MinerAddr
	if minerAddr == "" {
		minerAddr = fmt.Sprintf("0x%s", hex.EncodeToString(c.address))
	}

	block, err := c.blockchain.MineNewBlock(minerAddr)
	if err != nil {
		c.logSkipped(fmt.Sprintf("打包失败: %v", err), true)
		return
	}

	block.Header.Extra = []byte(currentAddr)

	if !c.blockchain.AddBlock(block) {
		c.logSkipped("加链失败", true)
		return
	}

	log.Printf("[PoA] 成功 打包 块 #%s", block.Header.Number.String())
	c.resetSkipLogState() // 重置跳过状态
	go c.broadcastBlock(block)
}

// 智能跳过日志打印
func (c *Core) logSkipped(reason string, forceLog bool) {
	now := time.Now()

	// 总是记录错误类日志
	if forceLog {
		log.Printf("[PoA] %s，跳过打包", reason)
		c.lastSkipReason = reason
		c.lastSkipLogTime = now
		return
	}

	// 检查是否需要打印无交易日志
	shouldLog := false
	if reason != c.lastSkipReason {
		shouldLog = true // 原因变化时打印
	} else if now.Sub(c.lastSkipLogTime) >= c.skipLogInterval {
		shouldLog = true // 超过时间间隔时打印
	}

	if shouldLog {
		log.Printf("[PoA] %s，跳过打包", reason)
		c.lastSkipReason = reason
		c.lastSkipLogTime = now
	}
}

// 重置跳过日志状态
func (c *Core) resetSkipLogState() {
	c.lastSkipReason = ""
	c.lastSkipLogTime = time.Now()
}

// 广播区块
func (c *Core) broadcastBlock(block *core.Block) {
	encodedBlock, err := block.EncodeBlock()
	if err != nil {
		log.Printf("[PoA] 区块编码失败: %v", err)
		return
	}

	msg := &p2p.Message{
		Type: p2p.BlockMessage,
		Data: encodedBlock,
	}
	encodedMsg, err := msg.Encode()
	if err != nil {
		log.Printf("[PoA] 消息编码失败: %v", err)
		return
	}

	c.p2pNode.BroadcastMessage(string(encodedMsg))
}

// 处理收到的 P2P 消息
func (c *Core) HandleP2PMessage(data []byte) {
	msg := &p2p.Message{}
	if err := msg.Decode(data); err != nil {
		log.Printf("[PoA] P2P 消息解码失败: %v", err)
		return
	}

	if msg.Type != p2p.BlockMessage {
		log.Printf("[PoA] 忽略非区块消息类型: %d", msg.Type)
		return
	}

	block := &core.Block{}
	if err := block.DecodeBlock(msg.Data); err != nil {
		log.Printf("[PoA] Block 解码失败: %v", err)
		return
	}

	c.OnReceiveBlock(block)
}

// 收到区块后的处理逻辑
func (c *Core) OnReceiveBlock(block *core.Block) {
	if !c.ValidateBlock(block) {
		log.Printf("[PoA] 区块验证失败，高度 #%d", block.Header.Number.Uint64())
		return
	}

	if !c.blockchain.AddBlock(block) {
		log.Printf("[PoA] 区块加链失败，高度 #%d", block.Header.Number.Uint64())
		return
	}

	log.Printf("[PoA] 成功同步区块 #%d", block.Header.Number.Uint64())
}

// 验证收到区块是否合法（轮值）
func (c *Core) ValidateBlock(block *core.Block) bool {
	height := block.Header.Number.Uint64()
	validatorIndex := height % uint64(len(c.ValSet))
	expected := c.ValSet[validatorIndex]
	actual := string(block.Header.Extra)

	if expected != actual {
		log.Printf("[PoA] 验证失败，预期验证者: 0x%s, 实际: 0x%s", expected, actual)
		return false
	}
	return true
}
