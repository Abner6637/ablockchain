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
	View     *big.Int
	Sequence *big.Int

	Preprepare *bft.Preprepare
	Prepares   *messageSet
	Commits    *messageSet
	lock       *sync.RWMutex
}

func NewConsensusState(view *big.Int, sequence *big.Int, preprepare *bft.Preprepare) *consensusState {
	return &consensusState{
		View:       view,
		Sequence:   sequence,
		Preprepare: preprepare,
		Prepares:   NewMessageSet(),
		Commits:    NewMessageSet(),
		lock:       new(sync.RWMutex),
	}
}

/*
// 结构化输出结构体时，使得未导出（小写开头）字段view和sequence能够转换为string类打印出
func (s *consensusState) String() string {
	return fmt.Sprintf("{view: %v, sequence: %v}", s.view, s.sequence)
}
*/

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

	return s.View
}

func (s *consensusState) getSequence() *big.Int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.Sequence
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
		View:     new(big.Int).Set(s.View),
		Sequence: new(big.Int).Set(s.Sequence),
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
	binary.Write(&buf, binary.BigEndian, s.Preprepare.Request.Time)

	digest := crypto.GlobalHashAlgorithm.Hash(buf.Bytes())

	return &bft.Commit{
		View:     new(big.Int).Set(s.View),
		Sequence: new(big.Int).Set(s.Sequence),
		Digest:   digest,
	}
}

func (s *consensusState) addCommit(msg *pbfttypes.Message) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	s.Commits.messages[string(msg.Address)] = msg
}

func (s *consensusState) getViewChange(newView *big.Int) *bft.ViewChange {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.Preprepare == nil {
		return nil
	}

	digest := s.Preprepare.Request.HashReqeust()

	return &bft.ViewChange{
		View:     newView,
		Sequence: new(big.Int).Set(s.Sequence),
		Digest:   digest,
	}
}

func (s *consensusState) getNewView(newView *big.Int) *bft.NewView {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.Preprepare == nil {
		return nil
	}

	return &bft.NewView{
		View:     newView,
		Sequence: new(big.Int).Set(s.Sequence),
	}
}
