package raftcore

import (
	"ablockchain/cli"
	rafttypes "ablockchain/consensus/bft/raft/types"
	"ablockchain/crypto"
	"ablockchain/p2p"
	"crypto/ecdsa"
	"encoding/hex"
	"log"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

type Core struct {
	p2pNode *p2p.Node

	privateKey *ecdsa.PrivateKey
	address    []byte

	state       rafttypes.State
	currentTerm uint64
	votedFor    []byte

	peers    []string
	addrToID map[string]string

	electionTimer *time.Timer

	wg sync.WaitGroup
	mu sync.Mutex

	votesReceived int
}

func NewCore(cfg *cli.Config, p2pNode *p2p.Node) *Core {
	address := crypto.PubkeyToAddress(p2pNode.PrivateKey.PublicKey).Bytes()

	var peers []string
	for _, s := range cfg.ValSet {
		s = strings.TrimPrefix(s, "0x")
		strBytes, _ := hex.DecodeString(s)
		peers = append(peers, string(strBytes))
	}
	sort.Strings(peers)

	return &Core{
		p2pNode:     p2pNode,
		privateKey:  p2pNode.PrivateKey,
		address:     address,
		state:       rafttypes.StateFollower,
		currentTerm: 0,
		votedFor:    nil,
		peers:       peers,
		addrToID:    make(map[string]string),
	}
}

func (c *Core) Start() error {
	log.Printf("[Raft] 节点启动，地址: 0x%x", c.address)

	c.startElectionTimer()

	c.wg.Add(1)
	go c.handleEvents()

	return nil
}

func (c *Core) Stop() error {
	if c.electionTimer != nil {
		c.electionTimer.Stop()
	}
	c.wg.Wait()
	log.Println("[Raft] 停止")

	return nil
}

func (c *Core) startElectionTimer() {
	timeout := time.Duration(5+rand.Intn(5)) * time.Second
	log.Printf("[Raft] 启动选举超时: %v", timeout)

	c.electionTimer = time.AfterFunc(timeout, func() {
		log.Printf("[Raft] 超时触发选举")
		c.startElection()
	})
}

func (c *Core) startElection() {

	c.votesReceived = 1

	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentTerm += 1
	c.state = rafttypes.StateCandidate
	c.votedFor = c.address

	log.Printf("[Raft] 成为候选人，Term: %d", c.currentTerm)

	for _, peer := range c.peers {
		if peer == string(c.address) {
			continue
		}
		msg := &rafttypes.Message{
			Type:        rafttypes.MsgRequestVote,
			Term:        c.currentTerm,
			CandidateID: c.address,
			Address:     c.address,
		}
		c.Unicast(msg, peer)
	}
}

func (c *Core) Unicast(msg *rafttypes.Message, address string) error {
	payload, _ := msg.Encode()
	p2pMsg := &p2p.Message{
		Type: p2p.ConsensusMessage,
		Data: payload,
	}
	encoded, _ := p2pMsg.Encode()

	if id, ok := c.addrToID[address]; ok {
		peerID, _ := peer.Decode(id)
		c.p2pNode.SendMessage(peerID, string(encoded))
	}
	return nil
}
