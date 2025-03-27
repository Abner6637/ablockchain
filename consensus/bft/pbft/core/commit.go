package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"ablockchain/event"
	"log"
)

func (c *Core) HandleCommit(msg *pbfttypes.Message) error {
	log.Printf("开始处理Commit消息：%+v", msg)
	// 验证消息签名
	if err := VerifySignature(msg); err != nil {
		return err
	}

	var commit *bft.Commit
	err := msg.Decode(&commit)
	if err != nil {
		return err
	}

	if c.consensusState.Preprepare != nil {
		if commit.View.Cmp(c.consensusState.View) != 0 || commit.Sequence.Cmp(c.consensusState.Sequence) != 0 {
			return nil
		}
	} else {
		return nil
	}

	c.consensusState.addCommit(msg)
	log.Printf("目前已存储的Commit消息数目：%d", len(c.consensusState.Commits.messages))

	if len(c.consensusState.Commits.messages) >= 2*c.ByzantineSize+1 && c.state.Cmp(pbfttypes.StateCommitted) < 0 {
		c.setState(pbfttypes.StateCommitted)

		block, err := c.consensusState.getBlock()
		if err != nil {
			return err
		}

		/*
			block, ok := c.pendingBlocks[string(blockHash)]
			if !ok {
				log.Fatalf("There is no block in pending according to the given hash!")
			}
		*/

		log.Printf("共识完成，consensus发布ConsensusFinish事件（告知区块链处理共识完成的区块）")
		event.Bus.Publish("ConsensusFinish", block)

		log.Printf("共识完成，consensus发布FinalCommitedBlock事件（告知共识模块准备处理下一个区块）")
		event.Bus.Publish("FinalCommitedBlock", block)

	}

	return nil
}

func (c *Core) SendCommit() error {
	var msg pbfttypes.Message
	msg.Code = pbfttypes.MsgCommit

	commit := c.consensusState.getCommit()

	encodedCommit, err := pbfttypes.Encode(commit)
	if err != nil {
		return err
	}

	// 打印使用
	log.Printf("生成Commit：%+v", commit)

	msg.Msg = encodedCommit
	msg.Address = c.address
	msg.Signature, err = c.SignMessage(&msg)
	if err != nil {
		return err
	}

	c.consensusState.addCommit(&msg)

	log.Printf("广播Commit消息：%+v", msg)
	c.Broadcast(&msg)

	return nil
}
