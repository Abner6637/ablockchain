package hotstuffcore

import (
	"ablockchain/consensus/bft"
	hotstufftypes "ablockchain/consensus/bft/hotstuff/types"
	"ablockchain/event"
	"log"
)

func (c *Core) HandleCommit2(msg *hotstufftypes.Message) error {
	if c.IsPrimary() {
		log.Printf("主节点，丢弃Commit2消息：%+v", msg)
		return nil
	}

	log.Printf("开始处理Commit2消息：%+v", msg)

	// 验证聚合签名
	if err := c.VerifyAggregateSignature(msg); err != nil {
		return err
	}

	var commit *bft.Commit
	err := msg.Decode(&commit)
	if err != nil {
		return err
	}

	if c.consensusState.Preprepare != nil {
		if commit.View.Cmp(c.consensusState.View) != 0 || commit.Sequence.Cmp(c.consensusState.Sequence) != 0 {
			log.Printf("Commit2消息的View或Sequence不匹配，丢弃消息")
			return nil
		}
	} else {
		return nil
	}

	c.consensusState.addCommit(msg)
	log.Printf("目前已存储的Commit消息数目：%d", len(c.consensusState.Commits.messages))

	if c.state.Cmp(hotstufftypes.StateCommitted) < 0 {
		c.setState(hotstufftypes.StateCommitted)

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

func (c *Core) SendCommit2() error {
	var msg hotstufftypes.Message
	msg.Code = hotstufftypes.MsgCommit2

	commit := c.consensusState.getCommit()

	encodedCommit, err := hotstufftypes.Encode(commit)
	if err != nil {
		return err
	}

	// 打印使用
	log.Printf("生成Commit2：%+v", commit)

	msg.Msg = encodedCommit
	msg.Address = c.address

	aggSig, valList, err := c.SignAggregate(c.consensusState.Commits)
	if err != nil {
		log.Printf("聚合签名失败：%v", err)
	}

	msg.Additional = valList
	msg.Signature = aggSig

	if err != nil {
		return err
	}

	c.consensusState.addCommit(&msg)

	log.Printf("广播Commit2消息：%+v", msg)
	c.Broadcast(&msg)

	return nil
}
