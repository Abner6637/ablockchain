package pbftcore

import (
	"ablockchain/cli"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
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
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
)

type Core struct {
	p2pNode *p2p.Node

	privateKey *ecdsa.PrivateKey
	address    []byte

	consensusState *consensusState

	state            pbfttypes.State
	curCommitedBlock *core.Block

	pendingRequests   *prque.Prque
	pendingRequestsMu *sync.Mutex

	events       []event.EventSubscription
	timeoutEvent event.EventSubscription

	Primary []byte // 主节点采用Valset经过sort后的顺序；如，view 0的时候，主节点为ValSet[0]; view 1的时候，主节点为ValSet[1]
	ValSet  []string

	ByzantineSize int
	NodeSize      int

	ViewChanges     map[uint64]*messageSet // key: view; value: messageSet
	viewChangeTimer *time.Timer
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
		state:             pbfttypes.StateAcceptRequest,
		privateKey:        p2pNode.PrivateKey,
		address:           crypto.PubkeyToAddress(p2pNode.PrivateKey.PublicKey).Bytes(),
		ByzantineSize:     (cfg.ConsensusNum - 1) / 3,
		NodeSize:          cfg.ConsensusNum,
		pendingRequests:   prque.New(),
		pendingRequestsMu: new(sync.Mutex),
		curCommitedBlock:  &core.Block{},
		ValSet:            valSet,
		Primary:           address, // 初始化为自身的地址（后续初始化过程中会更改）
		ViewChanges:       make(map[uint64]*messageSet),
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

	c.HandleEvents()

	c.StartNewProcess(big.NewInt(0))

	// log.Printf("是否为主节点：%v", c.IsPrimary())

	return nil
}

func (c *Core) Stop() error {
	event.Bus.Publish("ConsensusStop", true)

	c.UnSubcribeEvents()

	log.Println("PBFT stop")

	return nil
}

func (c *Core) Address() []byte {
	return c.address
}

func (c *Core) AddVal(address string) {

}

func (c *Core) Broadcast(msg *pbfttypes.Message) error {
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

	c.p2pNode.BroadcastMessage(string(encodedP2PMsg))

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

func (c *Core) setState(state pbfttypes.State) {
	c.state = state
	log.Printf("共识state变更为：%d", c.state)
}

func (c *Core) addViewChange(view *big.Int, msg *pbfttypes.Message) {
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
			c.consensusState = NewConsensusState(c.consensusState.getView(), big.NewInt(int64(c.curCommitedBlock.Header.Number)+1), nil)

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
		c.newViewChangeTimer()
	}
}

// 返回msg.Signature和err
func (c *Core) Sign(data []byte) ([]byte, error) {
	hashData := crypto.GlobalHashAlgorithm.Hash(data)
	return crypto.Sign(hashData, c.privateKey)
}

func (c *Core) SignMessage(msg *pbfttypes.Message) ([]byte, error) {
	// Sign message
	data, err := msg.PayloadNoSig()
	if err != nil {
		return nil, err
	}
	msg.Signature, err = c.Sign(data)
	if err != nil {
		return nil, err
	}

	return msg.Signature, err
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

func VerifySignature(msg *pbfttypes.Message) error {
	payloadNoSig, err := msg.PayloadNoSig()
	if err != nil {
		return err
	}

	signerAddress, err := GetSignatureAddress(payloadNoSig, msg.Signature)

	// 比较签名地址和消息中的地址参数（即发送消息的地址）是否一致
	if !bytes.Equal(signerAddress.Bytes(), msg.Address) {
		return errors.New("invaid signer")
	}
	return nil
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
