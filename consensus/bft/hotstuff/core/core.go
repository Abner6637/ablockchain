package hotstuffcore

import (
	"ablockchain/cli"
	hotstufftypes "ablockchain/consensus/bft/hotstuff/types"
	"ablockchain/core"
	"ablockchain/crypto"
	"ablockchain/event"
	"ablockchain/p2p"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"log"
	"math"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/libp2p/go-libp2p/core/peer"
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
)

type Core struct {
	p2pNode *p2p.Node

	privateKey   *ecdsa.PrivateKey
	address      []byte
	addrToID     map[string]string           // 地址到peerID的映射
	addrToBLSPub map[string]crypto.BLSPubKey // 地址到BLS公钥的映射

	consensusState *consensusState

	state            hotstufftypes.State
	curCommitedBlock *core.Block

	pendingRequests   *prque.Prque
	pendingRequestsMu *sync.Mutex

	events       []event.EventSubscription
	timeoutEvent event.EventSubscription
	helloEvent   event.EventSubscription

	Primary []byte // 主节点采用Valset经过sort后的顺序；如，view 0的时候，主节点为ValSet[0]; view 1的时候，主节点为ValSet[1]
	ValSet  []string

	ByzantineSize int
	NodeSize      int

	ViewChanges     map[uint64]*messageSet // key: view; value: messageSet
	viewChangeTimer *time.Timer

	wg *sync.WaitGroup
}

func NewCore(cfg *cli.Config, p2pNode *p2p.Node) *Core {
	address := crypto.PubkeyToAddress(p2pNode.PrivateKey.PublicKey).Bytes()

	var valSet []string
	if len(cfg.ValSet) != 0 {
		for _, str := range cfg.ValSet {
			str = strings.TrimPrefix(str, "0x")
			strBytes, err := hex.DecodeString(str)
			if err != nil {
				log.Fatalf("地址转换失败: %v", err)
			}
			valSet = append(valSet, string(strBytes))
		}
	} else {
		valSet = append(valSet, string(address))
	}

	sort.Strings(valSet)

	log.Printf("ValSet: %v", valSet)
	log.Printf("HexValSet: %x", valSet)

	return &Core{
		p2pNode:           p2pNode,
		state:             hotstufftypes.StateAcceptRequest,
		privateKey:        p2pNode.PrivateKey,
		address:           crypto.PubkeyToAddress(p2pNode.PrivateKey.PublicKey).Bytes(),
		addrToID:          make(map[string]string),
		addrToBLSPub:      make(map[string]crypto.BLSPubKey),
		ByzantineSize:     (cfg.ConsensusNum - 1) / 3,
		NodeSize:          cfg.ConsensusNum,
		pendingRequests:   prque.New(),
		pendingRequestsMu: new(sync.Mutex),
		curCommitedBlock:  &core.Block{},
		ValSet:            valSet,
		Primary:           address, // 初始化为自身的地址（后续初始化过程中会更改）
		ViewChanges:       make(map[uint64]*messageSet),
		wg:                &sync.WaitGroup{},
	}
}

func (c *Core) Start() error {

	log.Printf("start core-----------------")

	// log.Printf("consensusAddress: %s\n", string(c.Address()))
	log.Printf("HexAddress: 0x%x\n", c.Address())

	//hexAddress := fmt.Sprintf("0x%x", c.Address())
	// log.Printf("HexAddress2: %s\n", hexAddress)

	//hexAddress = strings.TrimPrefix(hexAddress, "0x")

	//addressBytes, err := hex.DecodeString(hexAddress)
	//if err != nil {
	//	log.Fatalf("地址转换失败: %v", err)
	//}

	//log.Printf("addressBytes: %s\n", string(addressBytes))
	//log.Printf("addressBytesToAddress: 0x%x\n", addressBytes)

	c.SubcribeEvents()

	c.wg.Add(1)
	go c.HandleEvents()

	c.StartNewProcess(big.NewInt(0))

	// log.Printf("是否为主节点：%v", c.IsPrimary())

	return nil
}

func (c *Core) Stop() error {
	event.Bus.Publish("ConsensusStop", true)

	c.wg.Wait()

	c.UnSubcribeEvents()

	log.Println("PBFT stop")

	return nil
}

func (c *Core) Address() []byte {
	return c.address
}

func (c *Core) AddVal(address string) {

}

func (c *Core) Broadcast(msg *hotstufftypes.Message) error {
	payload, err := msg.EncodeMessage()
	if err != nil {
		return err
	}

	p2pMsg := &p2p.Message{
		Type: p2p.ConsensusMessage,
		Data: payload,
	}
	encodedP2PMsg, err := p2pMsg.Encode()
	if err != nil {
		return err
	}

	if c.p2pNode != nil {
		c.p2pNode.BroadcastMessage(string(encodedP2PMsg))
	} else {
		return errors.New("不存在共识模块对应的p2pNode!")
	}

	return nil
}

func (c *Core) Unicast(msg *hotstufftypes.Message, address string) error {
	payload, err := msg.EncodeMessage()
	if err != nil {
		return err
	}

	p2pMsg := &p2p.Message{
		Type: p2p.ConsensusMessage,
		Data: payload,
	}
	encodedP2PMsg, err := p2pMsg.Encode()
	if err != nil {
		return err
	}

	if c.p2pNode != nil {
		if id, ok := c.addrToID[address]; ok {
			// 将string形式的id解码为peerID格式（string就是用base58编码的peerID）
			peerID, err := peer.Decode(id)
			if err != nil {
				log.Printf("peerID转换失败: %v", err)
				return err
			}
			c.p2pNode.SendMessage(peerID, string(encodedP2PMsg))
		} else {
			log.Printf("找不到地址 %s 对应的P2PID", address)
			return errors.New("地址不在共识节点列表中")
		}
	} else {
		return errors.New("不存在共识模块对应的p2pNode!")
	}

	return nil

}

func (c *Core) IsPrimary() bool {
	return bytes.Equal(c.Primary, c.address)
}

func (c *Core) PrimaryFromView(view *big.Int) string {
	res := new(big.Int)
	res.Mod(view, big.NewInt(int64(c.NodeSize)))
	return c.ValSet[int(res.Int64())]
}

func (c *Core) setState(state hotstufftypes.State) {
	c.state = state
	log.Printf("共识state变更为：%d", c.state)
}

func (c *Core) addViewChange(view *big.Int, msg *hotstufftypes.Message) {
	newView := view.Uint64()
	if c.ViewChanges[newView] == nil {
		c.ViewChanges[newView] = NewMessageSet()
	}
	c.ViewChanges[newView].messages[string(msg.Address)] = msg
}

func (c *Core) StartNewProcess(num *big.Int) {
	log.Printf("准备开始新一轮共识")
	if c.consensusState == nil { // 初始化
		c.consensusState = NewConsensusState(big.NewInt(0), big.NewInt(0), nil)
	} else {
		if num.Cmp(big.NewInt(0)) == 0 { // 没有视图切换
			c.consensusState = NewConsensusState(c.consensusState.getView(), new(big.Int).Add(c.curCommitedBlock.Header.Number, big.NewInt(1)), nil)

			// 正常共识完一个request后（没有视图转换等），开始处理队列中的下一个request（如果有的话）
			c.ProcessRequest()
		} else { // 发生视图切换
			c.consensusState = NewConsensusState(num, c.consensusState.getSequence(), c.consensusState.Preprepare) // preprepare继续上一视图正在进行的
		}
	}
	log.Printf("新的共识状态：%+v", c.consensusState)

	// 更新主节点（主节点随view编号变动）
	c.Primary = []byte(c.PrimaryFromView(c.consensusState.getView()))
	log.Printf("当前主节点地址: 0x%x", string(c.Primary))
	log.Printf("自己是否为主节点：%v", c.IsPrimary())

	// 当自身不为主节点时，启动viewchange计时器（因为该计时器用于主节点轮换，主节点自己不需要轮换自己）
	if !c.IsPrimary() {
		// c.newViewChangeTimer()
	}
}

// 返回msg.Signature和err
func (c *Core) Sign(data []byte) ([]byte, error) {
	blsPrivKey, err := crypto.ConvertEthToBLS(c.privateKey)
	if err != nil {
		return nil, err
	}

	blsSig, err := crypto.BLSSign(data, blsPrivKey)
	if err != nil {
		return nil, err
	}
	return blsSig.Serialize(), nil

}

// BLS签名
func (c *Core) SignMessage(msg *hotstufftypes.Message) ([]byte, error) {
	// Sign message
	data := msg.Msg

	blsSig, err := c.Sign(data)
	if err != nil {
		return nil, err
	}
	msg.Signature = blsSig

	return msg.Signature, nil
}

// 验证BLS签名
func (c *Core) VerifySignature(msg *hotstufftypes.Message) error {
	srcAddr := msg.Address
	srcPubKey, ok := c.addrToBLSPub[string(srcAddr)]
	if !ok {
		return errors.New("未找到对应的公钥")
	}

	var blsSig crypto.BLSSig
	blsSig.Deserialize(msg.Signature)
	ok = crypto.BLSVerify(msg.Msg, &srcPubKey, &blsSig)
	if !ok {
		return errors.New("验证签名失败")
	}
	return nil
}

// 签名聚合签名，返回序列化的聚合签名，编码后的验证者列表，error
func (c *Core) SignAggregate(messageSet *messageSet) ([]byte, []byte, error) {
	var signs []crypto.BLSSig // 签名列表
	var valLists []string     // 签名的验证者列表

	for addr, msg := range messageSet.messages {
		if msg.Code == hotstufftypes.MsgPrepare1 {
			c.consensusState.PrepareValList = append(c.consensusState.PrepareValList, addr)

			var blsSig crypto.BLSSig
			blsSig.Deserialize(msg.Signature)
			signs = append(signs, blsSig)

			valLists = append(valLists, addr)

		} else if msg.Code == hotstufftypes.MsgCommit1 {
			c.consensusState.CommitValList = append(c.consensusState.CommitValList, addr)

			var blsSig crypto.BLSSig
			blsSig.Deserialize(msg.Signature)
			signs = append(signs, blsSig)

			valLists = append(valLists, addr)
		}
	}

	log.Printf("验证者列表：%x", valLists)

	encodedValLists, _ := rlp.EncodeToBytes(valLists)

	// 对所有的签名进行聚合
	blsSig, err := crypto.AggregateSignatures(signs)
	if err != nil {
		return nil, nil, err
	}

	return blsSig.Serialize(), encodedValLists, nil
}

// 验证聚合签名
func (c *Core) VerifyAggregateSignature(msg *hotstufftypes.Message) error {
	var valLists []string
	rlp.DecodeBytes(msg.Additional, &valLists)
	if len(valLists) == 0 {
		return errors.New("缺失对应的验证者列表")
	}

	var valPubKeys []crypto.BLSPubKey
	for _, addr := range valLists {
		// 如果验证者地址是自身的话，则使用自身的bls公钥
		if addr == string(c.Address()) {
			blsPrivKey, _ := crypto.ConvertEthToBLS(c.privateKey)
			blsPubKey := blsPrivKey.GetPublicKey()
			valPubKeys = append(valPubKeys, *blsPubKey)
			continue
		}
		pubKey, ok := c.addrToBLSPub[addr]
		if !ok {
			return errors.New("未找到对应的公钥")
		}
		valPubKeys = append(valPubKeys, pubKey)
	}

	aggKey, _ := crypto.AggregatePublicKeys(valPubKeys)

	var blsSig crypto.BLSSig
	blsSig.Deserialize(msg.Signature)

	ok := crypto.VerifyAggregateSignature(aggKey, &blsSig, msg.Msg)
	if !ok {
		return errors.New("验证聚合签名失败")
	}
	return nil
}

// 通过未经哈希的原数据和签名得到签名所用的公钥，再通过公钥得到签名地址
func GetSignatureAddress(data []byte, sig []byte) (common.Address, error) {
	hashData := crypto.GlobalHashAlgorithm.Hash(data)

	pubkey, err := crypto.SigToPub(hashData, sig)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*pubkey), nil
}

func (c *Core) newViewChangeTimer() {
	if c.viewChangeTimer != nil {
		c.viewChangeTimer.Stop()
	}

	// 当timeout时间内未收到主节点的request时，发起viewchange
	// timeout会随着view的增加而逐渐增大（防止短时间内触发多个timeout事件）
	timeout := time.Duration(20 * time.Second)
	view := c.consensusState.getView().Uint64()
	if view > 0 {
		timeout += time.Duration(math.Pow(2, float64(view))) * time.Second
	}
	log.Printf("启动新的viewChangeTimer，timeout时间为：%v", timeout)
	log.Printf("--------------------------------------------------------------------")
	c.viewChangeTimer = time.AfterFunc(timeout, func() {
		event.Bus.Publish("TimeoutEvent", nil)
	})
}
