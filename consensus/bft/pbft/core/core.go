package pbftcore

import (
	"ablockchain/cli"
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"ablockchain/core"
	"ablockchain/crypto"
	"ablockchain/event"
	"ablockchain/p2p"
	"bytes"
	"crypto/ecdsa"
	"errors"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Core struct {
	p2pNode *p2p.Node

	privateKey *ecdsa.PrivateKey
	address    []byte

	consensusState *consensusState

	state            pbfttypes.State
	curCommitedBlock *core.Block

	pendingRequests map[string]*bft.Request

	events []event.EventSubscription

	Primary       string
	NodeSet       []string // 通过config注入
	ByzantineSize int
}

func NewCore(cfg *cli.Config, p2pNode *p2p.Node) *Core {
	return &Core{
		p2pNode:          p2pNode,
		state:            pbfttypes.StateAcceptRequest,
		privateKey:       p2pNode.PrivateKey,
		address:          crypto.PubkeyToAddress(p2pNode.PrivateKey.PublicKey).Bytes(),
		ByzantineSize:    (cfg.ConsensusNum - 1) / 3,
		pendingRequests:  make(map[string]*bft.Request),
		curCommitedBlock: &core.Block{},
	}
}

func (c *Core) Start() error {

	log.Printf("start core-----------------")

	c.SubcribeEvents()

	c.HandleEvents()

	c.StartNewProcess(big.NewInt(0))

	return nil
}

func (c *Core) Stop() error {
	event.Bus.Publish("ConsensusStop", true)

	c.UnSubcribeEvents()

	log.Println("PBFT stop")

	return nil
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
	return c.Primary == string(c.address)
}

func (c *Core) setState(state pbfttypes.State) {
	c.state = state
	log.Printf("共识state变更为：%d", c.state)
}

func (c *Core) StartNewProcess(num *big.Int) {
	log.Printf("准备开始新一轮共识")
	if c.consensusState == nil {
		c.consensusState = NewConsensusState(big.NewInt(0), big.NewInt(0), nil)
		log.Printf("新的共识状态：%+v", c.consensusState)
	} else {
		c.consensusState = NewConsensusState(c.consensusState.getView(), big.NewInt(int64(c.curCommitedBlock.Header.Number)+1), nil)
		log.Printf("更改共识状态：%+v", c.consensusState)
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
