package pow

import (
	"ablockchain/core"
	"ablockchain/crypto"
	"ablockchain/event"
	"ablockchain/p2p"
	"bytes"
	"fmt"
	"log"
	"math/big"
	"reflect"
	"time"
)

type ProofOfWork struct {
	p2pNode   *p2p.Node
	block     *core.Block
	target    *big.Int //用于判断hash前置0个数是否达到Difficulty要求
	running   bool     //用于停止挖矿
	stopcount bool     //挖矿途中收到新区块,停止当前计算
	events    []event.EventSubscription
}

func NewProofOfWork(p2pNode *p2p.Node) *ProofOfWork {
	return &ProofOfWork{
		p2pNode:   p2pNode,
		running:   false,
		stopcount: false,
	}
}

// 准备用于计算hash的数据
func (pow *ProofOfWork) prepareData(nonce uint64) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.Header.ParentHash,
			pow.block.Header.MerkleRoot,
			[]byte(fmt.Sprintf("%v", pow.block.Header.Time)),
			[]byte(fmt.Sprintf("%d", pow.block.Header.Difficulty)),
			[]byte(fmt.Sprintf("%d", nonce)),
		},
		[]byte{},
	)
	return data
}

// 共识的核心逻辑
func (pow *ProofOfWork) Run(block *core.Block) {
	var hashInt big.Int
	var hash []byte
	pow.stopcount = false
	nonce := uint64(0)
	maxNonce := uint64(9223372036854775807)
	pow.block = block
	target := big.NewInt(1)
	target.Lsh(target, uint(256-pow.block.Header.Difficulty*4))
	pow.target = target

	fmt.Printf("Mining the block \"%d\"\n", pow.block.Header.Number)
	for nonce < maxNonce {
		if pow.stopcount { //为监听到其他区块,继续挖矿
			fmt.Printf("\n 收到其他区块,停止计算...")
			break
		}
		data := pow.prepareData(nonce)
		hash = crypto.GlobalHashAlgorithm.Hash(data)
		hashInt.SetBytes(hash[:])
		//结束条件：hash小于target
		if hashInt.Cmp(pow.target) == -1 {
			fmt.Printf("\n hash: %x", hash)
			fmt.Printf("\n nonce: %d", nonce)
			break
		} else {
			nonce++
		}
		time.Sleep(time.Millisecond * 10)
	}
	fmt.Print("\n\n")
	if !pow.stopcount {
		block.Hash = hash[:]
		block.Header.Nonce = nonce
	} else { //返回非法区块
		block.Hash = nil
		block.Header.Nonce = 0
	}
}

// 验证hash
func Validate(block *core.Block) bool {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-block.Header.Difficulty*4))
	var hashInt big.Int
	data := bytes.Join(
		[][]byte{
			block.Header.ParentHash,
			block.Header.MerkleRoot,
			[]byte(fmt.Sprintf("%v", block.Header.Time)),
			[]byte(fmt.Sprintf("%d", block.Header.Difficulty)),
			[]byte(fmt.Sprintf("%d", block.Header.Nonce)),
		},
		[]byte{},
	)
	hash := crypto.GlobalHashAlgorithm.Hash(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(target) == -1

	return isValid
}

// 实现共识接口
func (pow *ProofOfWork) Start() error {
	if pow.running {
		fmt.Println("PoW 已经在运行")
		return nil
	}
	fmt.Println("\nPoW Start")

	pow.running = true
	pow.SubcribeEvents()
	pow.HandleEvents()
	return nil
}

func (pow *ProofOfWork) Stop() error {
	if !pow.running {
		fmt.Println("PoW 已经停止")
		return nil
	}
	pow.running = false
	event.Bus.Publish("ConsensusStop", true)
	fmt.Println("PoW stop")
	return nil
}

func (pow *ProofOfWork) SubcribeEvents() {
	pow.events = []event.EventSubscription{
		{
			Name:    "ConsensusStart",
			Channel: event.Bus.Subscribe("ConsensusStart"),
		},
		{
			Name:    "ConsensusStop",
			Channel: event.Bus.Subscribe("ConsensusStop"),
		},
		{
			Name:    "BlockMessage",
			Channel: event.Bus.Subscribe("BlockMessage"),
		},
	}
}

func (pow *ProofOfWork) HandleEvents() {
	go func() {
		// 准备反射监听参数
		cases := make([]reflect.SelectCase, len(pow.events))
		eventNames := make([]string, len(pow.events)) // 保存对应的名称

		for i, sub := range pow.events {
			cases[i] = reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(sub.Channel),
			}
			eventNames[i] = sub.Name // 建立索引与名称的映射
		}

		for {
			// 阻塞监听事件
			chosen, value, ok := reflect.Select(cases)

			// 处理通道关闭
			if !ok {
				log.Printf("事件通道关闭: %s", eventNames[chosen])
				return
			}

			// 获取事件数据
			eventData := value.Interface()

			// 根据事件名称路由处理
			switch eventNames[chosen] {
			case "ConsensusStart":
				block, ok := eventData.(*core.Block)
				if !ok {
					log.Fatal("转换失败: 事件数据不是 *core.Block 类型")
				}
				fmt.Println("PoW 收到共识事件，开始计算区块:", block.Header.Number)
				go func() {
					pow.Run(block)
					if Validate(block) {
						fmt.Println("验证通过，准备上链")
						encodedBlock, err := block.EncodeBlock()
						if err != nil {
							log.Fatal("区块编码失败:encodedBlock, err := block.EncodeBlock()")
						}
						p2pMsg := &p2p.Message{
							Type: p2p.BlockMessage,
							Data: encodedBlock,
						}
						encodedP2PMsg, err := p2pMsg.Encode()
						if err != nil {
							log.Fatal("消息编码失败:encodedP2PMsg, err := p2pMsg.Encode()")
						}
						//挖出区块后进行广播
						pow.p2pNode.BroadcastMessage(string(encodedP2PMsg))
						event.Bus.Publish("ConsensusFinish", block)
					} else {
						fmt.Println("丢弃非法区块")
					}
				}()

			case "ConsensusStop":
				if eventData == true {
					fmt.Println("\n结束监听")
					return
				}

			case "BlockMessage":
				fmt.Println("\n收到BlockMessage,开始验证")
				pow.stopcount = true
				block, ok := eventData.(*core.Block)
				if !ok {
					log.Fatal("转换失败: 事件数据不是 *core.Block 类型")
				}
				if Validate(block) {
					fmt.Println("\n外部区块合法")
					pow.stopcount = true //停止本轮计算
					event.Bus.Publish("ConsensusFinish", block)
				} else {
					fmt.Println("外部区块非法")
				}

			default:
				log.Printf("未知事件类型: %s", eventNames[chosen])
			}
		}
	}()
}
