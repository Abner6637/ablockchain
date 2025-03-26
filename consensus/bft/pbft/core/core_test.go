package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"ablockchain/core"
	"ablockchain/crypto"
	"ablockchain/p2p"
	"bytes"
	"crypto/ecdsa"
	"log"
	"math/big"
	"testing"
	"time"
)

func newTestBlock() *core.Block {
	return &core.Block{
		Header: &core.BlockHeader{
			ParentHash: []byte("0df9a8f4a2f2fc354c3c8aa5e837d4db137f20ccbf3d8336e4c95ac9d0e2943e"),
			MerkleRoot: []byte("1cdfdf5680f2a639732f6aae64a8b96c10a913b46c8fcd908c9eb95925979974"),
			Time:       uint64(time.Now().Unix()),
			Difficulty: 2,
			Nonce:      0,
			Number:     13,
		},
	}
}

func newP2PNode() *p2p.Node {
	return &p2p.Node{}
}

func NewTestCoreForSign(privateKey *ecdsa.PrivateKey) *Core {
	return &Core{
		privateKey: privateKey,
		address:    crypto.PubkeyToAddress(privateKey.PublicKey).Bytes(),
	}
}

func NewTestCore(p2pNode *p2p.Node) *Core {
	return &Core{
		p2pNode:    p2pNode,
		state:      pbfttypes.StateAcceptRequest,
		privateKey: p2pNode.PrivateKey,
		address:    crypto.PubkeyToAddress(p2pNode.PrivateKey.PublicKey).Bytes(),
	}
}

func TestProcess(t *testing.T) {
	//block := newTestBlock()
	//p2pnode := newP2PNode()
}

func TestSign(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	log.Printf("privateKey: %+v", privateKey)

	if err != nil {
		t.Fatalf("生成密钥失败：%v", err)
	}

	core := NewTestCoreForSign(privateKey)

	publicKey := privateKey.PublicKey
	log.Printf("publicKey: %+v", publicKey)

	comAddress := crypto.PubkeyToAddress(publicKey)

	address := comAddress.Bytes()
	log.Printf("address: %+v", address)

	var msg pbfttypes.Message
	msg.Code = pbfttypes.MsgPrepare
	prepare, err := pbfttypes.Encode(&bft.Prepare{
		View:     big.NewInt(0),
		Sequence: big.NewInt(0),
		Digest:   []byte{},
	})
	if err != nil {
		t.Fatalf("消息编码失败: %v", err)
	}
	msg.Msg = prepare
	msg.Address = address
	msg.Signature, err = core.SignMessage(&msg)
	if err != nil {
		t.Fatalf("签名消息失败：%v", err)
	}

	/*
		if err := VerifySignature(&msg); err != nil {
			t.Fatalf("验证消息失败：%v", err)
		}
	*/

	payloadNoSig, err := msg.PayloadNoSig()
	if err != nil {
		t.Fatalf("编码消息失败：%v", err)
	}

	signerAddress, err := GetSignatureAddress(payloadNoSig, msg.Signature)
	log.Printf("signer address: %+v", signerAddress.Bytes())

	if !bytes.Equal(signerAddress.Bytes(), msg.Address) {
		t.Fatalf("验证失败")
	} else {
		log.Printf("地址相同，验证成功")
	}
}
