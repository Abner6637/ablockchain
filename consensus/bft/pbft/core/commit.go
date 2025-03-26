package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"ablockchain/event"
	"log"
)

func (c *Core) HandleCommit(msg *pbfttypes.Message) error {
	// 验证消息签名
	if err := VerifySignature(msg); err != nil {
		return err
	}

	var commit *bft.Commit
	err := msg.Decode(&commit)
	if err != nil {
		return err
	}

	c.consensusState.addCommit(msg)

	if len(c.consensusState.Commits.messages) >= 2*c.ByzantineSize+1 {
		c.setState(pbfttypes.StateCommitted)

		blockHash := c.consensusState.getBlockHash()

		block, ok := c.pendingBlocks[string(blockHash)]
		if !ok {
			log.Fatalf("There is no block in pending according to the given hash!")
		}

		event.Bus.Publish("ConsensusFinish", block)
		event.Bus.Publish("FinalCommitedBlock", block)

	}

	return nil
}

func (c *Core) SendCommit() error {
	var msg pbfttypes.Message
	msg.Code = pbfttypes.MsgCommit
	commit, err := pbfttypes.Encode(c.consensusState.getCommit())
	if err != nil {
		return err
	}
	msg.Msg = commit
	msg.Address = c.address
	msg.Signature, err = c.SignMessage(&msg)
	if err != nil {
		return err
	}

	c.consensusState.addCommit(&msg)
	c.Broadcast(&msg)

	return nil
}
