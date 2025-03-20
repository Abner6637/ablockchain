package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
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
	// TODO: 两种消息集合的数据结构
	Prepares uint64
	Commits  uint64
	lock     *sync.RWMutex
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

func (s *consensusState) getPrepare() *bft.Prepare {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.Preprepare == nil {
		return nil
	}

	var buf bytes.Buffer
	buf.Write(s.Preprepare.Request.Msg)
	binary.Write(&buf, binary.BigEndian, uint32(s.Preprepare.Request.Time.Unix()))

	digest := crypto.GlobalHashAlgorithm.Hash(buf.Bytes())

	return &bft.Prepare{
		View:     new(big.Int).Set(s.view),
		Sequence: new(big.Int).Set(s.sequence),
		Digest:   digest,
	}
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
