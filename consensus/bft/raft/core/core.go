package raftcore

import (
	"ablockchain/cli"
	rafttypes "ablockchain/consensus/bft/raft/types"
	"ablockchain/core"
	"ablockchain/crypto"
	"ablockchain/event"
	"ablockchain/p2p"
	"crypto/ecdsa"
	"encoding/hex"
	"log"
	"math/big"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/libp2p/go-libp2p/core/peer"
)

type eventSub struct {
	Name    string
	Channel <-chan any
}

type Core struct {
	p2pNode    *p2p.Node
	privateKey *ecdsa.PrivateKey
	address    []byte
	peers      []string

	addrToID map[string]string // 以太坊地址到PeerID的映射

	currentTerm uint64
	votedFor    []byte
	role        string // "Leader", "Follower", "Candidate"
	leaderID    []byte

	electionTimer *time.Timer
	appendAcks    map[string]bool
	blockPending  *core.Block

	commitCh chan *core.Block
	wg       *sync.WaitGroup
	subs     []eventSub

	committedBlock *core.Block
}

func (c *Core) EncodeBlock(b *core.Block) []byte {
	encoded, err := rlp.EncodeToBytes(b)
	if err != nil {
		log.Printf("[Raft] Block RLP 编码失败: %v", err)
		return nil
	}
	return encoded
}

func DecodeBlock(data []byte) (*core.Block, error) {
	var blk core.Block
	err := rlp.DecodeBytes(data, &blk)
	return &blk, err
}

// NewCore 初始化时初始化映射
func NewCore(cfg *cli.Config, p2pNode *p2p.Node) *Core {
	address := crypto.PubkeyToAddress(p2pNode.PrivateKey.PublicKey).Bytes()

	var peers []string
	for _, str := range cfg.ValSet {
		str = strings.TrimPrefix(str, "0x")
		strBytes, err := hex.DecodeString(str)
		if err != nil {
			log.Fatalf("地址转换失败: %v", err)
		}
		//peers = append(peers, string(strBytes))
		peers = append(peers, hex.EncodeToString(strBytes)) // ✅ hex string 形式存储
	}

	sort.Strings(peers)

	return &Core{
		p2pNode:    p2pNode,
		privateKey: p2pNode.PrivateKey,
		address:    address,
		peers:      peers,
		role:       "Leader", // 默认先设为Leader方便调试
		commitCh:   make(chan *core.Block),
		appendAcks: make(map[string]bool),
		wg:         &sync.WaitGroup{},
		addrToID:   make(map[string]string), // 初始化映射表
	}
}

func (c *Core) Start() error {
	log.Printf("[Raft] Node starting, address: 0x%x", c.address)
	c.subscribeEvents()
	c.wg.Add(1)
	return nil
}

func (c *Core) Stop() error {
	event.Bus.Publish("ConsensusStop", true)
	c.wg.Wait()
	log.Println("[Raft] Node stopped")
	return nil
}

func (c *Core) Address() []byte {
	return c.address
}

func (c *Core) IsLeader() bool {
	return c.role == "Leader"
}

func (c *Core) subscribeEvents() {
	sub1 := event.Bus.Subscribe("ConsensusStart")
	go func() {
		for ev := range sub1 {
			if blk, ok := ev.(*core.Block); ok {
				c.handleConsensusStart(blk)
			}
		}
	}()
	c.subs = append(c.subs, eventSub{Name: "ConsensusStart", Channel: sub1})

	sub2 := event.Bus.Subscribe("MessageEvent")
	go func() {
		for ev := range sub2 {
			if msg, ok := ev.(*p2p.Message); ok {
				c.handleMessageEvent(msg)
			}
		}
	}()
	c.subs = append(c.subs, eventSub{Name: "Message", Channel: sub2})

	sub3 := event.Bus.Subscribe("FinalCommitedBlock")
	go func() {
		for ev := range sub3 {
			if blk, ok := ev.(*core.Block); ok {
				c.handleFinalBlock(blk)
			}
		}
	}()
	c.subs = append(c.subs, eventSub{Name: "FinalCommitedBlock", Channel: sub3})

	sub4 := event.Bus.Subscribe("ConsensusStop")
	c.subs = append(c.subs, eventSub{Name: "ConsensusStop", Channel: sub4})

	// 订阅 HelloEvent 用于维护地址映射
	subHello := event.Bus.Subscribe("HelloEvent")
	go func() {
		for ev := range subHello {
			if info, ok := ev.(p2p.EthPeerInfo); ok {
				addrHex := hex.EncodeToString([]byte(info.EthAddr)) // ✅ 保证统一 hex 格式
				c.addrToID[addrHex] = info.PeerID
				log.Printf("[Raft] 更新地址映射: %s -> %s", info.EthAddr, info.PeerID)
			}
		}
	}()
	c.subs = append(c.subs, eventSub{Name: "HelloEvent", Channel: subHello})
}

func (c *Core) HandleEvents() {
	log.Println("[Raft] 启动事件监听...")

	cases := make([]reflect.SelectCase, len(c.subs))
	eventNames := make([]string, len(c.subs))

	for i, sub := range c.subs {
		cases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(sub.Channel),
		}
		eventNames[i] = sub.Name
	}

	for {
		chosen, value, ok := reflect.Select(cases)
		if !ok {
			log.Printf("[Raft] 事件通道关闭: %s", eventNames[chosen])
			return
		}

		eventData := value.Interface()
		name := eventNames[chosen]

		switch name {
		case "ConsensusStart":
			log.Println("[Raft] 收到 ConsensusStart 事件")
			block, ok := eventData.(*core.Block)
			if !ok {
				log.Printf("[Raft] ConsensusStart 类型断言失败，实际类型: %T", eventData)
				break
			}
			c.blockPending = block
			c.StartElection()

		case "Message":
			log.Println("[Raft] 收到 Message 事件")
			msg, ok := eventData.(*p2p.Message)
			if !ok {
				log.Printf("[Raft] Message 类型断言失败，实际类型: %T", eventData)
				break
			}
			c.handleMessageEvent(msg)

		case "FinalCommitedBlock":
			log.Println("[Raft] 收到 FinalCommitedBlock 事件")
			block, ok := eventData.(*core.Block)
			if !ok {
				log.Printf("[Raft] FinalCommitedBlock 类型断言失败，实际类型: %T", eventData)
				break
			}
			c.committedBlock = block
			c.StartNewTerm(big.NewInt(0))

		case "ConsensusStop":
			log.Println("[Raft] 收到 ConsensusStop 事件，退出")
			return

		default:
			log.Printf("[Raft] 未知事件: %s", name)
		}
	}
}

func (c *Core) handleConsensusStart(block *core.Block) {
	if !c.IsLeader() {
		return
	}
	log.Println("[Raft] Leader received ConsensusStart event, broadcasting AppendEntries")
	c.blockPending = block
	c.appendAcks = make(map[string]bool)
	c.broadcastAppendEntries(block)
}

func (c *Core) handleMessageEvent(msg *p2p.Message) {
	var ae rafttypes.AppendEntries

	// 直接用 msg.Data 进行 RLP 解码，无需 base64 解码
	err := rlp.DecodeBytes(msg.Data, &ae)
	if err != nil {
		log.Printf("[Raft] RLP 解码 AppendEntries 失败: %v", err)
		return
	}

	// 解码 Block
	block, err := core.DecodeBlock(ae.Data)
	if err != nil {
		log.Printf("[Raft] 解码 Block 失败: %v", err)
		return
	}

	log.Printf("[Raft] 成功解析 AppendEntries，区块高度：%d", block.Header.Number.Uint64())

	// 构造并发送响应
	resp := rafttypes.AppendResp{
		Term:    ae.Term,
		From:    hex.EncodeToString(c.address),
		Success: true,
	}

	encodedResp, err := rlp.EncodeToBytes(resp)
	if err != nil {
		log.Printf("[Raft] 响应 RLP 编码失败: %v", err)
		return
	}

	// 直接发送 []byte，无需 base64 编码
	leaderID, err := peer.Decode(string(ae.LeaderID))
	if err != nil {
		log.Printf("[Raft] LeaderID 解码失败: %v", err)
		return
	}

	err = c.p2pNode.SendMessage(leaderID, string(encodedResp))

	if err != nil {
		log.Printf("[Raft] 回复 AppendResp 失败: %v", err)
	}
}

func (c *Core) handleFinalBlock(block *core.Block) {
	log.Printf("[Raft] Block committed at height %d\n", block.Header.Number)

}

func (c *Core) broadcastAppendEntries(block *core.Block) {
	msg := rafttypes.AppendEntries{
		Term:     c.currentTerm,
		LeaderID: c.address,
		Data:     c.EncodeBlock(block),
	}
	encodedBytes, err := rlp.EncodeToBytes(msg)
	if err != nil {
		log.Printf("[Raft] RLP 编码 AppendEntries 失败: %v", err)
		return
	}

	for _, addr := range c.peers {
		if addr == hex.EncodeToString(c.address) {
			continue
		}
		peerIDStr, ok := c.addrToID[addr]
		if !ok {
			log.Printf("[Raft] 找不到地址对应的 PeerID: %s", addr)
			continue
		}
		peerID, err := peer.Decode(peerIDStr)
		if err != nil {
			log.Printf("[Raft] PeerID 解码失败: %v", err)
			continue
		}

		// 直接发送编码后的 []byte，无需 base64 编码
		err = c.p2pNode.SendMessage(peerID, string(encodedBytes))

		if err != nil {
			log.Printf("[Raft] 发送 AppendEntries 到 %s 失败: %v", peerIDStr, err)
		} else {
			log.Printf("[Raft] AppendEntries 已发送给 %s", peerIDStr)
		}
	}
}

func (c *Core) majority() int {
	return len(c.peers)/2 + 1
}

func (c *Core) countAcks() int {
	count := 0
	for _, ack := range c.appendAcks {
		if ack {
			count++
		}
	}
	return count
}

func (c *Core) StartElection() {
	log.Println("[Raft] StartElection() 被调用（未实现）")
	// TODO: 添加选举逻辑
}

func (c *Core) StartNewTerm(newTerm *big.Int) {
	log.Printf("[Raft] StartNewTerm() 被调用，新任期: %d", newTerm)
	// TODO: 添加换任期逻辑
}
