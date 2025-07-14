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

// 定义常量
const sealInterval = 30 * time.Second

// Core 定义 PoA 共识核心
type Core struct {
	p2pNode    *p2p.Node
	blockchain *core.Blockchain
	privateKey *ecdsa.PrivateKey
	address    []byte
	ValSet     []string
	wg         *sync.WaitGroup
}

// 创建共识核心
func NewCore(cfg *cli.Config, p2pNode *p2p.Node, bc *core.Blockchain) *Core {
	address := crypto.PubkeyToAddress(p2pNode.PrivateKey.PublicKey).Bytes()
	// 统一地址格式为小写无0x前缀的十六进制
	normalizeAddr := func(addr string) string {
		if strings.HasPrefix(addr, "0x") {
			return addr[2:]
		}
		return addr
	}
	// 如果配置中有验证者列表，则使用配置中的验证者，否则使用当前节点地址作为验证者
	var valSet []string
	if len(cfg.ValSet) > 0 {
		for _, addr := range cfg.ValSet {
			valSet = append(valSet, normalizeAddr(addr))
		}
	} else {
		valSet = append(valSet, hex.EncodeToString(address))
	}
	log.Printf("[PoA] 配置的验证者集合: %v", valSet)
	log.Printf("[PoA] 当前节点地址: %s", hex.EncodeToString(address))
	return &Core{
		p2pNode:    p2pNode,
		blockchain: bc,
		privateKey: p2pNode.PrivateKey,
		address:    address,
		ValSet:     valSet,
		wg:         &sync.WaitGroup{},
	}
}

// 启动共识
func (c *Core) Start() error {
	log.Printf("[PoA] 启动, 我的地址: 0x%x", c.Address())
	c.wg.Add(1)
	go c.loopSeal()
	return nil
}

// 停止共识
func (c *Core) Stop() error {
	c.wg.Wait()
	log.Println("[PoA] 停止")
	return nil
}

// 返回自己地址
func (c *Core) Address() []byte {
	return c.address
}

// 轮值定时器，定期尝试出块
func (c *Core) loopSeal() {
	defer c.wg.Done()

	ticker := time.NewTicker(sealInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.trySeal()
	}
}

// 出块尝试
func (c *Core) trySeal() {
	// 获取当前区块高度
	currentBlock := c.blockchain.CurrentBlock()
	if currentBlock == nil {
		log.Println("[PoA] 无法获取当前区块")
		return
	}
	height := c.blockchain.CurrentBlock().Header.Number.Uint64()

	// 轮值计算：根据高度计算当前验证者
	validatorIndex := height % uint64(len(c.ValSet))
	expected := c.ValSet[validatorIndex]
	currentAddr := hex.EncodeToString(c.address)
	log.Printf("[PoA] 当前区块高度: %d, 轮到验证者: %s", height, expected)
	// 如果当前不是轮到本节点进行出块，则跳过
	if expected != currentAddr {
		log.Printf("[PoA] 当前轮到验证者: 0x%s, 跳过", expected)
		return
	}

	// 当前节点需要出块
	log.Printf("[PoA] 当前轮到我 打包 块，高度 #%d", height+1)

	// 获取矿工地址
	minerAddr := c.blockchain.MinerAddr
	if minerAddr == "" {
		minerAddr = fmt.Sprintf("0x%s", hex.EncodeToString(c.address))
	}

	// 创建新区块
	block, err := c.blockchain.MineNewBlock(minerAddr)
	if err != nil {
		log.Println("[PoA] 没有可打包交易")
		return
	}

	// 标记出块人
	block.Header.Extra = c.address

	// 尝试将新区块添加到区块链中
	if !c.blockchain.AddBlock(block) {
		log.Println("[PoA] 加链失败")
		return
	}

	// 记录出块成功
	log.Printf("[PoA] 成功 打包 块 #%s", block.Header.Number.String())

	// 使用 Goroutine 异步广播新区块
	go c.broadcastBlock(block)
}

// 广播新区块
func (c *Core) broadcastBlock(block *core.Block) {
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

	// 广播消息
	c.p2pNode.BroadcastMessage(string(encodedP2PMsg))
}
