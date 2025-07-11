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
	"sync"
	"time"
)

// Core 定义 PoA 共识核心
type Core struct {
	p2pNode *p2p.Node

	blockchain *core.Blockchain

	privateKey *ecdsa.PrivateKey
	address    []byte

	ValSet []string
	wg     *sync.WaitGroup
}

// 创建共识核心
func NewCore(cfg *cli.Config, p2pNode *p2p.Node, bc *core.Blockchain) *Core {
	address := crypto.PubkeyToAddress(p2pNode.PrivateKey.PublicKey).Bytes()

	var valSet []string
	if len(cfg.ValSet) != 0 {
		valSet = cfg.ValSet
	} else {
		valSet = append(valSet, string(address))
	}

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

// 轮值定时器
func (c *Core) loopSeal() {
	defer c.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.trySeal()
		}
	}
}

// 出块尝试
func (c *Core) trySeal() {
	height := c.blockchain.CurrentBlock().Header.Number.Uint64()

	// 轮值计算
	validatorIndex := height % uint64(len(c.ValSet))
	expected := c.ValSet[validatorIndex]

	if expected != string(c.address) {
		log.Printf("[PoA] 当前轮到验证者: 0x%x, 跳过", expected)
		return
	}

	log.Printf("[PoA] 当前轮到我 Seal 块，高度 #%d", height+1)

	//minerAddr := fmt.Sprintf("0x%s", hex.EncodeToString(c.address))

	minerAddr := c.blockchain.MinerAddr
	if minerAddr == "" {
		// fallback 使用节点地址
		minerAddr = fmt.Sprintf("0x%s", hex.EncodeToString(c.address))
	}

	block, err := c.blockchain.MineNewBlock(minerAddr)
	if err != nil {
		log.Println("[PoA] 没有可打包交易")
		return
	}

	// 标记出块人
	block.Header.Extra = c.address

	if !c.blockchain.AddBlock(block) {
		log.Println("[PoA] 加链失败")
		return
	}

	log.Printf("[PoA] 成功 Seal 块 #%s", block.Header.Number.String())

	// 广播
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
}
