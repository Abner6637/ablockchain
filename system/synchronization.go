package system

import (
	"ablockchain/core"
	"ablockchain/event"
	"ablockchain/p2p"
	"bytes"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/rlp"
)

func ListenSynchronization(sys *System) {
	getSnapshotMessage := event.Bus.Subscribe("GetSnapshotMessage")
	snapshotMessage := event.Bus.Subscribe("SnapshotMessage")
	getBlocksMessage := event.Bus.Subscribe("GetBlocksMessage")
	blocksMessage := event.Bus.Subscribe("BlocksMessage")

	go func() {
		for {
			select {
			case msg := <-getSnapshotMessage:
				pid, ok := msg.(string)
				if !ok {
					log.Fatal("转换失败: 事件数据不是 string")
					continue
				}
				fmt.Printf("\n收到GetSnapshotMessage, PeerID: %s", pid)
				snapshotsmsg := &core.SnapshotMessage{
					PeerID:   sys.p2pNode.ID,
					Snapshot: sys.blockChain.CreateSnapshot(),
				}
				encodedgetSnapshotsmsg, err := snapshotsmsg.EncodeSnapshotMessage()
				if err != nil {
					log.Fatal("编码失败")
				}
				sys.p2pNode.SendP2PMessage(pid, p2p.SnapshotMessage, encodedgetSnapshotsmsg)

			case msg := <-snapshotMessage:
				data, ok := msg.(core.SnapshotMessage)
				if !ok {
					log.Fatal("转换失败data, ok := msg.(map[string]any)")
					continue
				}
				fmt.Printf("\n收到SnapshotMessage, PeerID: %s", data.PeerID)
				data.Snapshot.PrintSnapshot()
				switch sys.blockChain.CurBlockNum.Cmp(data.Snapshot.Number) {
				case -1: // 区块数量落后，请求最新区块
					getBlocksmsg := &core.GetBlocksMessage{
						PeerID: sys.p2pNode.ID,
						From:   new(big.Int).Add(sys.blockChain.CurBlockNum, big.NewInt(1)),
						To:     data.Snapshot.Number,
					}
					encodedgetBlocksmsg, err := getBlocksmsg.EncodeGetBlocksMessage()
					if err != nil {
						log.Fatal("编码失败")
					}
					sys.p2pNode.SendP2PMessage(data.PeerID, p2p.GetBlocksMessage, encodedgetBlocksmsg)

				case 0: // 等高，验证hash和root（是否分叉）
					if bytes.Equal(sys.blockChain.CurrentBlockHash, data.Snapshot.Hash) && bytes.Equal(sys.blockChain.StateRoot, data.Snapshot.Root) {
						fmt.Println("\n 同步完成")
					} else {
						fmt.Println("\n 出现分叉") //TODO
					}

				case 1: // 区块数量领先
					fmt.Println("\n 无法与高度更低的节点进行同步")
				}

			case msg := <-getBlocksMessage:
				data, ok := msg.(core.GetBlocksMessage)
				if !ok {
					log.Fatal("转换失败")
					continue
				}
				fmt.Printf("\n收到GetBlocksMessage, PeerID:%s, From:%s, To:%s", data.PeerID, data.From, data.To)
				// 打包blocks并发送
				bloks, err := sys.blockChain.GetBlocksInRange(data.From, data.To)
				if err != nil {
					log.Fatal("读取区块链失败")
					continue
				}
				blocksmsg := &core.BlocksMessage{
					PeerID: sys.p2pNode.ID,
					Blocks: bloks,
				}
				encodedblocksmsg, err := blocksmsg.EncodeBlocksMessage()
				if err != nil {
					log.Fatal("编码失败")
				}
				sys.p2pNode.SendP2PMessage(data.PeerID, p2p.BlocksMessage, encodedblocksmsg)

			case msg := <-blocksMessage:
				data, ok := msg.(core.BlocksMessage)
				if !ok {
					log.Fatal("转换失败")
					continue
				}
				fmt.Printf("\n收到BlocksMessage, PeerID:%s, 区块数量:%d", data.PeerID, len(data.Blocks))
				for _, block := range data.Blocks {
					if sys.isValidBlock(block) {
						sys.blockChain.AddBlock(block)
					} else {
						fmt.Println("区块不合法，同步中断！")
						continue
					}
				}
			}
		}
	}()
}

func (sys *System) isValidBlock(block *core.Block) bool {
	if !bytes.Equal(block.Hash, block.Header.BlockHash()) {
		fmt.Println("区块哈希错误\n")
		return false
	}
	if !bytes.Equal(sys.blockChain.CurrentBlockHash, block.Header.ParentHash) {
		fmt.Println("区块ParentHash 与 sys.blockChain.CurrentBlockHash不相等\n")
		return false
	}
	if sys.blockChain.CurBlockNum != new(big.Int).Add(block.Header.Number, big.NewInt(1)) {
		fmt.Println("区块序号不连续\n")
		return false
	}
	return true
}

func (c *Commander) handleSyc(addr string) {
	if len(c.sys.p2pNode.Peers()) < 1 {
		fmt.Println("错误: 未连接任何节点")
		return
	}
	encodedGetSnapshotmsg, err := rlp.EncodeToBytes(c.sys.p2pNode.ID)
	if err != nil {
		log.Fatal("RLP encoding failed:", err)
		return
	}
	// 发送给第一个连接的节点（实际应用可扩展选择机制）
	targetPeer := c.sys.p2pNode.Peers()[0]
	p2pMsg := &p2p.Message{
		Type: p2p.GetSnapshotMessage,
		Data: encodedGetSnapshotmsg,
	}
	encodedMsg, err := p2pMsg.Encode()
	if err != nil {
		fmt.Errorf("消息编码失败: %w", err)
		return
	}
	if err := c.sys.p2pNode.SendMessage(targetPeer, string(encodedMsg)); err != nil {
		fmt.Errorf("发送消息失败: %w", err)
		return
	}
}
