package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"ablockchain/core"
	"ablockchain/crypto"
	"bytes"
	"encoding/binary"
	"math/big"
	"sync"
)

type consensusState struct {
	view     *big.Int
	sequence *big.Int

	state pbfttypes.State

	Preprepare *bft.Preprepare
	Prepares   *messageSet // 记录每个Address对应的消息哈希
	Commits    *messageSet
	lock       *sync.RWMutex
}

func (s *consensusState) getBlock() (*core.Block, error) {
	req := s.Preprepare.Request
	block, err := core.DecodeBlock(req.Msg)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (s *consensusState) getView() *big.Int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.view
}

func (s *consensusState) getSequence() *big.Int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.sequence
}

func (s *consensusState) setPreprepare(preprepare *bft.Preprepare) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	s.Preprepare = preprepare
}

func (s *consensusState) getPrepare() *bft.Prepare {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.Preprepare == nil {
		return nil
	}

	digest := s.Preprepare.Request.HashReqeust()

	return &bft.Prepare{
		View:     new(big.Int).Set(s.view),
		Sequence: new(big.Int).Set(s.sequence),
		Digest:   digest,
	}
}

func (s *consensusState) addPrepare(msg *pbfttypes.Message) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	s.Prepares.messages[string(msg.Address)] = msg
}

func (s *consensusState) getCommit() *bft.Commit {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.Preprepare == nil {
		return nil
	}

	var buf bytes.Buffer
	buf.Write(s.Preprepare.Request.Msg)
	binary.Write(&buf, binary.BigEndian, uint32(s.Preprepare.Request.Time.Unix()))

	digest := crypto.GlobalHashAlgorithm.Hash(buf.Bytes())

	return &bft.Commit{
		View:     new(big.Int).Set(s.view),
		Sequence: new(big.Int).Set(s.sequence),
		Digest:   digest,
	}
}

func (s *consensusState) addCommit(msg *pbfttypes.Message) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	s.Commits.messages[string(msg.Address)] = msg
}

func (s *consensusState) setState(state pbfttypes.State) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	s.state = state
}
