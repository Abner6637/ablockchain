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

// package poacore

// import (
// 	"ablockchain/cli"
// 	"ablockchain/core"
// 	"ablockchain/crypto"
// 	"ablockchain/p2p"
// 	"crypto/ecdsa"
// 	"encoding/hex"
// 	"fmt"
// 	"log"
// 	"strings"
// 	"sync"
// 	"time"
// )

// // 定义常量
// const sealInterval = 30 * time.Second

// // Core 定义 PoA 共识核心
// type Core struct {
// 	p2pNode    *p2p.Node
// 	blockchain *core.Blockchain
// 	privateKey *ecdsa.PrivateKey
// 	address    []byte
// 	ValSet     []string
// 	wg         *sync.WaitGroup
// }

// // 创建共识核心
// func NewCore(cfg *cli.Config, p2pNode *p2p.Node, bc *core.Blockchain) *Core {
// 	address := crypto.PubkeyToAddress(p2pNode.PrivateKey.PublicKey).Bytes()
// 	// 统一地址格式为小写无0x前缀的十六进制
// 	normalizeAddr := func(addr string) string {
// 		if strings.HasPrefix(addr, "0x") {
// 			return addr[2:]
// 		}
// 		return addr
// 	}
// 	// 如果配置中有验证者列表，则使用配置中的验证者，否则使用当前节点地址作为验证者
// 	var valSet []string
// 	if len(cfg.ValSet) > 0 {
// 		for _, addr := range cfg.ValSet {
// 			valSet = append(valSet, normalizeAddr(addr))
// 		}
// 	} else {
// 		valSet = append(valSet, hex.EncodeToString(address))
// 	}
// 	log.Printf("[PoA] 配置的验证者集合: %v", valSet)
// 	log.Printf("[PoA] 当前节点地址: %s", hex.EncodeToString(address))
// 	return &Core{
// 		p2pNode:    p2pNode,
// 		blockchain: bc,
// 		privateKey: p2pNode.PrivateKey,
// 		address:    address,
// 		ValSet:     valSet,
// 		wg:         &sync.WaitGroup{},
// 	}
// }

// // 启动共识
// func (c *Core) Start() error {
// 	log.Printf("[PoA] 启动, 我的地址: 0x%x", c.Address())
// 	c.wg.Add(1)
// 	go c.loopSeal()
// 	return nil
// }

// // 停止共识
// func (c *Core) Stop() error {
// 	c.wg.Wait()
// 	log.Println("[PoA] 停止")
// 	return nil
// }

// // 返回自己地址
// func (c *Core) Address() []byte {
// 	return c.address
// }

// // 轮值定时器，定期尝试出块
// func (c *Core) loopSeal() {
// 	defer c.wg.Done()

// 	ticker := time.NewTicker(sealInterval)
// 	defer ticker.Stop()

// 	for range ticker.C {
// 		c.trySeal()
// 	}
// }

// // 出块尝试
// func (c *Core) trySeal() {
// 	// 获取当前区块高度
// 	currentBlock := c.blockchain.CurrentBlock()
// 	if currentBlock == nil {
// 		log.Println("[PoA] 无法获取当前区块")
// 		return
// 	}
// 	height := c.blockchain.CurrentBlock().Header.Number.Uint64()

// 	// 使用时间轮换验证者 (每30秒轮换一次)
// 	epoch := time.Now().Unix() / int64(sealInterval.Seconds())
// 	validatorIndex := epoch % int64(len(c.ValSet))
// 	expected := c.ValSet[validatorIndex]

// 	currentAddr := hex.EncodeToString(c.address)
// 	log.Printf("[PoA] 当前区块高度: %d, 轮到验证者: %s", height, expected)

// 	// 如果当前不是轮到本节点进行出块，则跳过
// 	if expected != currentAddr {
// 		log.Printf("[PoA] 当前轮到验证者: 0x%s, 跳过", expected)
// 		return
// 	}

// 	// 当前节点需要出块
// 	log.Printf("[PoA] 当前轮到我 打包 块，高度 #%d", height+1)

// 	// 获取矿工地址
// 	minerAddr := c.blockchain.MinerAddr
// 	if minerAddr == "" {
// 		minerAddr = fmt.Sprintf("0x%s", hex.EncodeToString(c.address))
// 	}

// 	// 创建新区块（允许空块）
// 	block, err := c.blockchain.MineNewBlock(minerAddr)
// 	if err != nil {
// 		log.Println("[PoA] 打包失败:", err)
// 		return
// 	}

// 	// 标记出块人
// 	block.Header.Extra = c.address

// 	// 尝试将新区块添加到区块链中
// 	if !c.blockchain.AddBlock(block) {
// 		log.Println("[PoA] 加链失败")
// 		return
// 	}

// 	// 记录出块成功
// 	log.Printf("[PoA] 成功 打包 块 #%s", block.Header.Number.String())

// 	// 使用 Goroutine 异步广播新区块
// 	go c.broadcastBlock(block)
// }

// // 广播新区块
// func (c *Core) broadcastBlock(block *core.Block) {
// 	encodedBlock, err := block.EncodeBlock()
// 	if err != nil {
// 		log.Printf("[PoA] 区块编码失败: %v", err)
// 		return
// 	}

// 	p2pMsg := &p2p.Message{
// 		Type: p2p.BlockMessage,
// 		Data: encodedBlock,
// 	}

// 	encodedP2PMsg, err := p2pMsg.Encode()
// 	if err != nil {
// 		log.Printf("[PoA] 消息编码失败: %v", err)
// 		return
// 	}

//		// 广播消息
//		c.p2pNode.BroadcastMessage(string(encodedP2PMsg))
//	}
// package poacore

// import (
// 	"ablockchain/cli"
// 	"ablockchain/core"
// 	"ablockchain/crypto"
// 	"ablockchain/p2p"
// 	"crypto/ecdsa"
// 	"encoding/hex"
// 	"fmt"
// 	"log"
// 	"strings"
// 	"sync"
// 	"time"
// )

// // 定义常量
// const sealInterval = 30 * time.Second

// // Core 定义 PoA 共识核心
// type Core struct {
// 	p2pNode    *p2p.Node
// 	blockchain *core.Blockchain
// 	privateKey *ecdsa.PrivateKey
// 	address    []byte
// 	ValSet     []string
// 	wg         *sync.WaitGroup
// }

// // 创建共识核心
// func NewCore(cfg *cli.Config, p2pNode *p2p.Node, bc *core.Blockchain) *Core {
// 	address := crypto.PubkeyToAddress(p2pNode.PrivateKey.PublicKey).Bytes()
// 	// 统一地址格式为小写无0x前缀的十六进制
// 	normalizeAddr := func(addr string) string {
// 		if strings.HasPrefix(addr, "0x") {
// 			return addr[2:]
// 		}
// 		return addr
// 	}
// 	// 如果配置中有验证者列表，则使用配置中的验证者，否则使用当前节点地址作为验证者
// 	var valSet []string
// 	if len(cfg.ValSet) > 0 {
// 		for _, addr := range cfg.ValSet {
// 			valSet = append(valSet, normalizeAddr(addr))
// 		}
// 	} else {
// 		valSet = append(valSet, hex.EncodeToString(address))
// 	}
// 	log.Printf("[PoA] 配置的验证者集合: %v", valSet)
// 	log.Printf("[PoA] 当前节点地址: %s", hex.EncodeToString(address))
// 	return &Core{
// 		p2pNode:    p2pNode,
// 		blockchain: bc,
// 		privateKey: p2pNode.PrivateKey,
// 		address:    address,
// 		ValSet:     valSet,
// 		wg:         &sync.WaitGroup{},
// 	}
// }

// // 启动共识
// func (c *Core) Start() error {
// 	log.Printf("[PoA] 启动, 我的地址: 0x%x", c.Address())
// 	c.wg.Add(1)
// 	go c.loopSeal()
// 	return nil
// }

// // 停止共识
// func (c *Core) Stop() error {
// 	c.wg.Wait()
// 	log.Println("[PoA] 停止")
// 	return nil
// }

// // 返回自己地址
// func (c *Core) Address() []byte {
// 	return c.address
// }

// // 轮值定时器，定期尝试出块
// func (c *Core) loopSeal() {
// 	defer c.wg.Done()

// 	ticker := time.NewTicker(sealInterval)
// 	defer ticker.Stop()

// 	for range ticker.C {
// 		c.trySeal()
// 	}
// }

// // 出块尝试
// func (c *Core) trySeal() {
// 	// 获取当前区块高度
// 	currentBlock := c.blockchain.CurrentBlock()
// 	if currentBlock == nil {
// 		log.Println("[PoA] 无法获取当前区块")
// 		return
// 	}
// 	height := c.blockchain.CurrentBlock().Header.Number.Uint64()

// 	// 使用时间轮换验证者 (每30秒轮换一次)
// 	epoch := time.Now().Unix() / int64(sealInterval.Seconds())
// 	validatorIndex := epoch % int64(len(c.ValSet))
// 	expected := c.ValSet[validatorIndex]

// 	currentAddr := hex.EncodeToString(c.address)
// 	log.Printf("[PoA] 当前区块高度: %d, 轮到验证者: %s", height, expected)

// 	// 如果当前不是轮到本节点进行出块，则跳过
// 	if expected != currentAddr {
// 		log.Printf("[PoA] 当前轮到验证者: 0x%s, 跳过", expected)
// 		return
// 	}

// 	// 当前节点需要出块
// 	log.Printf("[PoA] 当前轮到我 打包 块，高度 #%d", height+1)

// 	// 获取矿工地址
// 	minerAddr := c.blockchain.MinerAddr
// 	if minerAddr == "" {
// 		minerAddr = fmt.Sprintf("0x%s", hex.EncodeToString(c.address))
// 	}

// 	// 创建新区块（允许空块）
// 	block, err := c.blockchain.MineNewBlock(minerAddr)
// 	if err != nil {
// 		log.Println("[PoA] 打包失败:", err)
// 		return
// 	}

// 	// 标记出块人
// 	block.Header.Extra = c.address

// 	// 尝试将新区块添加到区块链中
// 	if !c.blockchain.AddBlock(block) {
// 		log.Println("[PoA] 加链失败")
// 		return
// 	}

// 	// 记录出块成功
// 	log.Printf("[PoA] 成功 打包 块 #%s", block.Header.Number.String())

// 	// 使用 Goroutine 异步广播新区块
// 	go c.broadcastBlock(block)
// }

// // 广播新区块
// func (c *Core) broadcastBlock(block *core.Block) {
// 	encodedBlock, err := block.EncodeBlock()
// 	if err != nil {
// 		log.Printf("[PoA] 区块编码失败: %v", err)
// 		return
// 	}

// 	p2pMsg := &p2p.Message{
// 		Type: p2p.BlockMessage,
// 		Data: encodedBlock,
// 	}

// 	encodedP2PMsg, err := p2pMsg.Encode()
// 	if err != nil {
// 		log.Printf("[PoA] 消息编码失败: %v", err)
// 		return
// 	}

// 	// 广播消息
// 	log.Printf("[PoA] 正在广播新区块: %s", block.Header.Number.String())
// 	c.p2pNode.BroadcastMessage(string(encodedP2PMsg))
// 	log.Println("[PoA] 区块广播完成")
// }

// // 广播交易（修改后）
// func (c *Core) broadcastTransaction(tx *core.Transaction) {
// 	encodedTx, err := tx.EncodeTransaction()
// 	if err != nil {
// 		log.Printf("[PoA] 交易编码失败: %v", err)
// 		return
// 	}

// 	// 确保 encodedTx 是 []byte 类型
// 	encodedTxBytes, ok := encodedTx.([]byte)
// 	if !ok {
// 		log.Printf("[PoA] 交易编码后的数据不是 []byte 类型")
// 		return
// 	}

// 	// 创建 p2p 消息并广播
// 	p2pMsg := &p2p.Message{
// 		Type: p2p.TransactionMessage,
// 		Data: encodedTxBytes, // 使用类型断言后的值
// 	}

// 	encodedP2PMsg, err := p2pMsg.Encode()
// 	if err != nil {
// 		log.Printf("[PoA] 消息编码失败: %v", err)
// 		return
// 	}

//		// 广播交易
//		c.p2pNode.BroadcastMessage(string(encodedP2PMsg))
//		log.Println("[PoA] 交易广播完成")
//	}
// package poacore

// import (
// 	"ablockchain/cli"
// 	"ablockchain/core"
// 	"ablockchain/crypto"
// 	"ablockchain/p2p"
// 	"crypto/ecdsa"
// 	"encoding/hex"
// 	"fmt"
// 	"log"
// 	"strings"
// 	"sync"
// 	"time"
// )

// const sealInterval = 30 * time.Second

// type Core struct {
// 	p2pNode    *p2p.Node
// 	blockchain *core.Blockchain
// 	privateKey *ecdsa.PrivateKey
// 	address    []byte
// 	ValSet     []string
// 	wg         *sync.WaitGroup
// }

// func NewCore(cfg *cli.Config, p2pNode *p2p.Node, bc *core.Blockchain) *Core {
// 	address := crypto.PubkeyToAddress(p2pNode.PrivateKey.PublicKey).Bytes()
// 	normalizeAddr := func(addr string) string {
// 		if strings.HasPrefix(addr, "0x") {
// 			return addr[2:]
// 		}
// 		return addr
// 	}

// 	var valSet []string
// 	if len(cfg.ValSet) > 0 {
// 		for _, addr := range cfg.ValSet {
// 			valSet = append(valSet, normalizeAddr(addr))
// 		}
// 	} else {
// 		valSet = append(valSet, hex.EncodeToString(address))
// 	}

// 	log.Printf("[PoA] 配置的验证者集合: %v", valSet)
// 	log.Printf("[PoA] 当前节点地址: %s", hex.EncodeToString(address))

// 	return &Core{
// 		p2pNode:    p2pNode,
// 		blockchain: bc,
// 		privateKey: p2pNode.PrivateKey,
// 		address:    address,
// 		ValSet:     valSet,
// 		wg:         &sync.WaitGroup{},
// 	}
// }

// func (c *Core) Start() error {
// 	log.Printf("[PoA] 启动, 我的地址: 0x%x", c.Address())
// 	c.wg.Add(1)
// 	go c.loopSeal()
// 	return nil
// }

// func (c *Core) Stop() error {
// 	c.wg.Wait()
// 	log.Println("[PoA] 停止")
// 	return nil
// }

// func (c *Core) Address() []byte {
// 	return c.address
// }

// func (c *Core) loopSeal() {
// 	defer c.wg.Done()
// 	ticker := time.NewTicker(sealInterval)
// 	defer ticker.Stop()
// 	for range ticker.C {
// 		c.trySeal()
// 	}
// }

// func (c *Core) trySeal() {
// 	currentBlock := c.blockchain.CurrentBlock()
// 	if currentBlock == nil {
// 		log.Println("[PoA] 无法获取当前区块")
// 		return
// 	}
// 	height := currentBlock.Header.Number.Uint64()
// 	epoch := time.Now().Unix() / int64(sealInterval.Seconds())
// 	validatorIndex := epoch % int64(len(c.ValSet))
// 	expected := c.ValSet[validatorIndex]
// 	currentAddr := hex.EncodeToString(c.address)

// 	log.Printf("[PoA] 当前区块高度: %d, 轮到验证者: %s", height, expected)

// 	if expected != currentAddr {
// 		log.Printf("[PoA] 当前轮到验证者: 0x%s, 跳过", expected)
// 		return
// 	}

// 	pendingTxs := c.blockchain.TxPool.Pending
// 	if len(pendingTxs) == 0 {
// 		log.Println("[PoA] 没有待处理交易，跳过打包")
// 		return
// 	}

// 	log.Printf("[PoA] 当前轮到我 打包 块，高度 #%d", height+1)

// 	minerAddr := c.blockchain.MinerAddr
// 	if minerAddr == "" {
// 		minerAddr = fmt.Sprintf("0x%s", hex.EncodeToString(c.address))
// 	}

// 	block, err := c.blockchain.MineNewBlock(minerAddr)
// 	if err != nil {
// 		log.Println("[PoA] 打包失败:", err)
// 		return
// 	}

// 	block.Header.Extra = c.address

// 	if !c.blockchain.AddBlock(block) {
// 		log.Println("[PoA] 加链失败")
// 		return
// 	}

// 	log.Printf("[PoA] 成功 打包 块 #%s", block.Header.Number.String())
// 	go c.broadcastBlock(block)
// }

// func (c *Core) broadcastBlock(block *core.Block) {
// 	encodedBlock, err := block.EncodeBlock()
// 	if err != nil {
// 		log.Printf("[PoA] 区块编码失败: %v", err)
// 		return
// 	}

// 	p2pMsg := &p2p.Message{
// 		Type: p2p.BlockMessage,
// 		Data: encodedBlock,
// 	}

// 	encodedP2PMsg, err := p2pMsg.Encode()
// 	if err != nil {
// 		log.Printf("[PoA] 消息编码失败: %v", err)
// 		return
// 	}

// 	log.Printf("[PoA] 正在广播新区块: %s", block.Header.Number.String())
// 	c.p2pNode.BroadcastMessage(string(encodedP2PMsg))
// 	log.Println("[PoA] 区块广播完成")
// }

// func (c *Core) broadcastTransaction(tx *core.Transaction) {
// 	encodedTx, err := tx.EncodeTransaction()
// 	if err != nil {
// 		log.Printf("[PoA] 交易编码失败: %v", err)
// 		return
// 	}

// 	txBytes, ok := encodedTx.([]byte)
// 	if !ok {
// 		log.Printf("[PoA] 交易编码结果不是 []byte")
// 		return
// 	}

// 	p2pMsg := &p2p.Message{
// 		Type: p2p.TransactionMessage,
// 		Data: txBytes,
// 	}

// 	encodedP2PMsg, err := p2pMsg.Encode()
// 	if err != nil {
// 		log.Printf("[PoA] 消息编码失败: %v", err)
// 		return
// 	}

// 	c.p2pNode.BroadcastMessage(string(encodedP2PMsg))
// 	log.Println("[PoA] 交易广播完成")
// }

//	func (c *Core) HandleMessage(msg *p2p.Message) {
//		switch msg.Type {
//		case p2p.TransactionMessage:
//			tx := &core.Transaction{}
//			err := tx.DecodeTransaction(msg.Data)
//			if err != nil {
//				log.Printf("[PoA] 接收交易解码失败: %v", err)
//				return
//			}
//			log.Println("[PoA] 接收到广播交易，加入交易池")
//			err = c.blockchain.TxPool.AddTransaction(tx)
//			if err != nil {
//				log.Printf("[PoA] 添加交易失败: %v", err)
//			}
//		}
//	}
// package poacore

// import (
// 	"ablockchain/core"
// 	"ablockchain/crypto"
// 	"ablockchain/event"
// 	"ablockchain/p2p"
// 	"crypto/ecdsa"
// 	"encoding/hex"

// 	"log"
// 	"sort"

// 	"sync"
// 	"time"
// )

// const consensusInterval = 30 * time.Second

// type Core struct {
// 	p2pNode    *p2p.Node
// 	blockchain *core.Blockchain
// 	privateKey *ecdsa.PrivateKey
// 	address    []byte
// 	valSet     []string
// 	eventCh    chan interface{}
// 	running    bool
// 	wg         sync.WaitGroup
// }

// func NewCore(p2pNode *p2p.Node, bc *core.Blockchain, valSet []string) *Core {
// 	address := crypto.PubkeyToAddress(p2pNode.PrivateKey.PublicKey).Bytes()
// 	addrHex := hex.EncodeToString(address)

// 	if len(valSet) == 0 {
// 		valSet = append(valSet, addrHex)
// 	}

// 	sort.Strings(valSet)
// 	log.Printf("[PoA] 初始化验证者集合: %v", valSet)

// 	return &Core{
// 		p2pNode:    p2pNode,
// 		blockchain: bc,
// 		privateKey: p2pNode.PrivateKey,
// 		address:    address,
// 		valSet:     valSet,
// 		eventCh:    make(chan interface{}, 10),
// 	}
// }

// func (c *Core) Start() error {
// 	if c.running {
// 		return nil
// 	}
// 	c.running = true
// 	log.Printf("[PoA] 共识启动，节点地址: 0x%x", c.address)
// 	c.subscribe()
// 	c.wg.Add(2)
// 	go c.consensusLoop()
// 	go c.messageLoop()
// 	return nil
// }

// func (c *Core) Stop() error {
// 	if !c.running {
// 		return nil
// 	}
// 	c.running = false
// 	close(c.eventCh)
// 	c.wg.Wait()
// 	log.Println("[PoA] 共识停止")
// 	return nil
// }

// func (c *Core) subscribe() {
// 	event.Bus.SubscribeAsync("IncomingBlock", c.eventCh)
// 	event.Bus.SubscribeAsync("IncomingTx", c.eventCh)
// 	event.Bus.SubscribeAsync("Shutdown", c.eventCh)
// }

// func (c *Core) consensusLoop() {
// 	defer c.wg.Done()
// 	ticker := time.NewTicker(consensusInterval)
// 	defer ticker.Stop()

// 	for c.running {
// 		select {
// 		case <-ticker.C:
// 			c.tryPropose()
// 		}
// 	}
// }

// func (c *Core) messageLoop() {
// 	defer c.wg.Done()
// 	for msg := range c.eventCh {
// 		switch m := msg.(type) {
// 		case *core.Block:
// 			c.handleBlock(m)
// 		case *core.Transaction:
// 			c.handleTransaction(m)
// 		case string:
// 			if m == "Shutdown" {
// 				log.Println("[PoA] 收到 Shutdown 消息，退出")
// 				return
// 			}
// 		}
// 	}
// }

// func (c *Core) tryPropose() {
// 	pending := c.blockchain.TxPool.Pending
// 	if len(pending) == 0 {
// 		log.Println("[PoA] 无交易，跳过出块")
// 		return
// 	}

// 	height := c.blockchain.CurrentBlock().Header.Number.Uint64() + 1
// 	index := height % uint64(len(c.valSet))
// 	expected := c.valSet[index]
// 	myAddr := hex.EncodeToString(c.address)

// 	if expected != myAddr {
// 		log.Printf("[PoA] 当前不轮到我出块 (期望: %s)", expected)
// 		return
// 	}

// 	block, err := c.blockchain.MineNewBlock("0x" + myAddr)
// 	if err != nil {
// 		log.Printf("[PoA] 出块失败: %v", err)
// 		return
// 	}
// 	block.Header.Extra = c.address
// 	c.broadcastBlock(block)
// 	c.blockchain.AddBlock(block)
// }

// func (c *Core) handleBlock(block *core.Block) {
// 	height := block.Header.Number.Uint64()
// 	index := height % uint64(len(c.valSet))
// 	expected := c.valSet[index]
// 	actual := hex.EncodeToString(block.Header.Extra)
// 	if expected != actual {
// 		log.Printf("[PoA] 区块验证失败 (期望: %s, 实际: %s)", expected, actual)
// 		return
// 	}
// 	log.Printf("[PoA] 接收到合法区块 #%d", height)
// 	c.blockchain.AddBlock(block)
// }

// func (c *Core) handleTransaction(tx *core.Transaction) {
// 	err := c.blockchain.TxPool.AddTransaction(tx)
// 	if err != nil {
// 		log.Printf("[PoA] 添加交易失败: %v", err)
// 		return
// 	}
// 	log.Println("[PoA] 交易加入交易池")
// }

// func (c *Core) HandleMessage(msg *p2p.Message) {
// 	switch msg.Type {
// 	case p2p.BlockMessage:
// 		block := &core.Block{}
// 		if err := block.DecodeBlock(msg.Data); err != nil {
// 			log.Printf("[PoA] 区块解码失败: %v", err)
// 			return
// 		}
// 		event.Bus.Publish("IncomingBlock", block)
// 	case p2p.TransactionMessage:
// 		tx, err := core.DecodeTx(msg.Data)
// 		if err != nil {
// 			log.Printf("[PoA] 交易解码失败: %v", err)
// 			return
// 		}
// 		event.Bus.Publish("IncomingTx", tx)
// 	}
// }

//	func (c *Core) broadcastBlock(block *core.Block) {
//		encoded, err := block.EncodeBlock()
//		if err != nil {
//			log.Printf("[PoA] 区块编码失败: %v", err)
//			return
//		}
//		msg := &p2p.Message{
//			Type: p2p.BlockMessage,
//			Data: encoded,
//		}
//		send, err := msg.Encode()
//		if err != nil {
//			log.Printf("[PoA] 消息编码失败: %v", err)
//			return
//		}
//		c.p2pNode.BroadcastMessage(string(send))
//		log.Printf("[PoA] 区块 #%s 广播完成", block.Header.Number.String())
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
	p2pNode    *p2p.Node
	blockchain *core.Blockchain
	privateKey *ecdsa.PrivateKey
	address    []byte
	ValSet     []string
	wg         *sync.WaitGroup
	stopChan   chan struct{}
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
		p2pNode:    p2pNode,
		blockchain: bc,
		privateKey: p2pNode.PrivateKey,
		address:    address,
		ValSet:     valSet,
		wg:         &sync.WaitGroup{},
		stopChan:   make(chan struct{}),
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
		log.Println("[PoA] 无法获取当前区块")
		return
	}
	height := currentBlock.Header.Number.Uint64()
	nextHeight := height + 1

	validatorIndex := nextHeight % uint64(len(c.ValSet))
	expected := c.ValSet[validatorIndex]
	currentAddr := hex.EncodeToString(c.address)

	if expected != currentAddr {
		log.Printf("[PoA] 当前轮到验证者: 0x%s, 跳过", expected)
		return
	}

	if c.blockchain.TxPool.PendingSize() == 0 {
		log.Println("[PoA] 没有待处理交易，跳过打包")
		return
	}

	log.Printf("[PoA] 当前轮到我 打包 块 #%d", nextHeight)

	minerAddr := c.blockchain.MinerAddr
	if minerAddr == "" {
		minerAddr = fmt.Sprintf("0x%s", hex.EncodeToString(c.address))
	}

	block, err := c.blockchain.MineNewBlock(minerAddr)
	if err != nil {
		log.Println("[PoA] 打包失败:", err)
		return
	}

	block.Header.Extra = []byte(currentAddr)

	if !c.blockchain.AddBlock(block) {
		log.Println("[PoA] 加链失败")
		return
	}

	log.Printf("[PoA] 成功 打包 块 #%s", block.Header.Number.String())
	go c.broadcastBlock(block)
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
