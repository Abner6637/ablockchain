package raftcore

import (
	rafttypes "ablockchain/consensus/bft/raft/types"
	"ablockchain/event"
	"ablockchain/p2p"
	"log"
)

func (c *Core) handleEvents() {
	defer c.wg.Done()

	// 共识消息事件
	consensusCh := event.Bus.Subscribe("ConsensusMessage")
	// 节点握手事件
	helloCh := event.Bus.Subscribe("HelloEvent")

	for {
		select {
		case data := <-consensusCh:
			c.handleMessage(data)
		case data := <-helloCh:
			c.handleHello(data)
		}
	}
}

func (c *Core) handleMessage(data interface{}) {
	raw, ok := data.([]byte)
	if !ok {
		log.Println("[Raft] 消息格式错误")
		return
	}

	msg, err := rafttypes.DecodeMessage(raw)
	if err != nil {
		log.Println("[Raft] 消息解码错误")
		return
	}

	switch msg.Type {
	case rafttypes.MsgRequestVote:
		c.handleRequestVote(msg)
	case rafttypes.MsgVoteResponse:
		c.handleVoteResponse(msg)
	}
}

func (c *Core) handleHello(data interface{}) {
	info, ok := data.(p2p.EthPeerInfo)
	if !ok {
		log.Println("[Raft] HelloEvent 格式错误")
		return
	}

	c.addrToID[info.EthAddr] = info.PeerID
	log.Printf("[Raft] 记录 EthAddr -> PeerID: %s -> %s", info.EthAddr, info.PeerID)
}

func (c *Core) handleVoteResponse(msg *rafttypes.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if msg.Term != c.currentTerm {
		return
	}

	if msg.Success {
		c.votesReceived++
		log.Printf("[Raft] 收到投票回复，当前票数: %d", c.votesReceived)

		if c.votesReceived > len(c.peers)/2 && c.state == rafttypes.StateCandidate {
			log.Printf("[Raft] 赢得选举，成为 Leader")
			c.state = rafttypes.StateLeader
			// 成为 Leader 后，停止选举定时器
			if c.electionTimer != nil {
				c.electionTimer.Stop()
			}
			// TODO: 开启心跳定时器
		}
	}
}

func (c *Core) handleRequestVote(msg *rafttypes.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if msg.Term < c.currentTerm {
		return
	}

	// 简化: 接受投票
	c.currentTerm = msg.Term
	c.votedFor = msg.CandidateID

	// 回复
	reply := &rafttypes.Message{
		Type:    rafttypes.MsgVoteResponse,
		Term:    c.currentTerm,
		Success: true,
		Address: c.address,
	}
	c.Unicast(reply, string(msg.CandidateID))
	log.Printf("[Raft] 投票给: %x", msg.CandidateID)
}
