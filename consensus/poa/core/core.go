package poacore

import (
	"ablockchain/cli"
	poatypes "ablockchain/consensus/poa/types"
	"ablockchain/core"

	//"ablockchain/crypto"
	"ablockchain/p2p"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Core struct {
	p2pNode    *p2p.Node
	blockChain *core.Blockchain

	privateKey *ecdsa.PrivateKey
	address    []byte

	addrToID map[string]string
	ValSet   []string
	Primary  []byte

	NodeSize      int
	ByzantineSize int

	wg *sync.WaitGroup

	// 新增: Prepare 和 Commit 收集
	prepareSet map[string]map[string]bool // blockHash -> addr -> bool
	commitSet  map[string]map[string]bool
}

// 创建核心
func NewCore(cfg *cli.Config, p2pNode *p2p.Node, bc *core.Blockchain) *Core {
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

	log.Printf("ValSet: %x", valSet)

	return &Core{
		p2pNode:       p2pNode,
		blockChain:    bc,
		privateKey:    p2pNode.PrivateKey,
		address:       address,
		addrToID:      make(map[string]string),
		ValSet:        valSet,
		Primary:       address,
		NodeSize:      cfg.ConsensusNum,
		ByzantineSize: (cfg.ConsensusNum - 1) / 3,
		wg:            &sync.WaitGroup{},
		prepareSet:    make(map[string]map[string]bool),
		commitSet:     make(map[string]map[string]bool),
	}
}

func (c *Core) Start() error {
	log.Println("PoA core started")
	log.Printf("HexAddress: 0x%x", c.Address())
	c.Primary = []byte(c.ValSet[0])
	log.Printf("当前主节点: 0x%x", c.Primary)

	if c.IsPrimary() {
		c.ProcessRequest()
	}
	return nil
}

func (c *Core) Stop() error {
	log.Println("PoA core stopped")
	return nil
}

func (c *Core) Address() []byte {
	return c.address
}

func (c *Core) IsPrimary() bool {
	return bytes.Equal(c.Primary, c.address)
}

func (c *Core) IsValidator(address []byte) bool {
	for _, v := range c.ValSet {
		if bytes.Equal([]byte(v), address) {
			return true
		}
	}
	return false
}

func (c *Core) Broadcast(msg *poatypes.Message) error {
	payload, err := msg.EncodeMessage()
	if err != nil {
		return err
	}
	p2pMsg := &p2p.Message{
		Type: p2p.ConsensusMessage,
		Data: payload,
	}
	encoded, err := p2pMsg.Encode()
	if err != nil {
		return err
	}
	c.p2pNode.BroadcastMessage(string(encoded))
	return nil
}

func (c *Core) Unicast(msg *poatypes.Message, address string) error {
	payload, err := msg.EncodeMessage()
	if err != nil {
		return err
	}
	p2pMsg := &p2p.Message{
		Type: p2p.ConsensusMessage,
		Data: payload,
	}
	encoded, err := p2pMsg.Encode()
	if err != nil {
		return err
	}

	if id, ok := c.addrToID[address]; ok {
		peerID, err := peer.Decode(id)
		if err != nil {
			return err
		}
		c.p2pNode.SendMessage(peerID, string(encoded))
	} else {
		return errors.New("Unicast: address not found")
	}
	return nil
}

// 主节点生成初始请求
func (c *Core) ProcessRequest() {
	log.Println("主节点生成 Preprepare")

	preprepare := &poatypes.Message{
		Code:    poatypes.MsgPreprepare,
		Msg:     []byte("SomeBlockHashOrData"), // 实际使用请替换为真实区块 hash 或数据
		Address: c.Address(),
	}
	c.SignMessage(preprepare)
	c.Broadcast(preprepare)

	prepare := &poatypes.Message{
		Code:    poatypes.MsgPrepare,
		Msg:     preprepare.Msg,
		Address: c.Address(),
	}
	c.SignMessage(prepare)
	log.Println("主节点本地处理 Prepare")
	c.HandleMessage(prepare)
}

// 核心: 处理共识消息
func (c *Core) HandleMessage(msg *poatypes.Message) {
	data, _ := msg.PayloadNoSig()
	signer, err := GetSignatureAddress(data, msg.Signature)
	if err != nil {
		log.Println("签名恢复失败:", err)
		return
	}

	if !c.IsValidator(signer.Bytes()) {
		log.Println("收到非法验证者消息，忽略")
		return
	}

	hash := string(msg.Msg)

	switch msg.Code {
	case poatypes.MsgPreprepare:
		log.Println("收到 Preprepare")
		if c.prepareSet[hash] == nil {
			c.prepareSet[hash] = make(map[string]bool)
		}
		c.prepareSet[hash][string(c.address)] = true

		prepare := &poatypes.Message{
			Code:    poatypes.MsgPrepare,
			Msg:     msg.Msg,
			Address: c.Address(),
		}
		c.SignMessage(prepare)
		c.Broadcast(prepare)

	case poatypes.MsgPrepare:
		if c.prepareSet[hash] == nil {
			c.prepareSet[hash] = make(map[string]bool)
		}
		c.prepareSet[hash][string(msg.Address)] = true

		log.Printf("Prepare 当前计数: %d", len(c.prepareSet[hash]))
		if len(c.prepareSet[hash]) >= 2*c.ByzantineSize+1 {
			log.Println("Prepare 阈值达标，生成 Commit")
			commit := &poatypes.Message{
				Code:    poatypes.MsgCommit,
				Msg:     msg.Msg,
				Address: c.Address(),
			}
			c.SignMessage(commit)
			c.Broadcast(commit)
		}

	case poatypes.MsgCommit:
		if c.commitSet[hash] == nil {
			c.commitSet[hash] = make(map[string]bool)
		}
		c.commitSet[hash][string(msg.Address)] = true

		log.Printf("Commit 当前计数: %d", len(c.commitSet[hash]))
		if len(c.commitSet[hash]) >= 2*c.ByzantineSize+1 {
			log.Println("Commit 阈值达标，正式提交区块")
			block, err := c.blockChain.MineNewBlock("0x" + hex.EncodeToString(c.Address()))
			if err == nil {
				c.blockChain.AddBlock(block)
			} else {
				log.Println("打包区块失败:", err)
			}
		}
	}
}

/*
	 func (c *Core) Sign(data []byte) ([]byte, error) {
		hash := crypto.GlobalHashAlgorithm.Hash(data)
		return crypto.Sign(hash, c.privateKey)
	}
*/
func (c *Core) Sign(data []byte) ([]byte, error) {
	hash := crypto.Keccak256(data)
	return crypto.Sign(hash, c.privateKey)
}

func (c *Core) SignMessage(msg *poatypes.Message) ([]byte, error) {
	data, err := msg.PayloadNoSig()
	if err != nil {
		return nil, err
	}
	sig, err := c.Sign(data)
	if err != nil {
		return nil, err
	}
	msg.Signature = sig
	return sig, nil
}

//func GetSignatureAddress(data []byte, sig []byte)

func GetSignatureAddress(data []byte, sig []byte) (common.Address, error) {
	hash := crypto.Keccak256(data)
	pubKeyBytes, err := crypto.Ecrecover(hash, sig)
	if err != nil {
		return common.Address{}, err
	}
	pubKey, err := crypto.UnmarshalPubkey(pubKeyBytes)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*pubKey), nil
}
