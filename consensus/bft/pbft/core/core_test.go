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

func newTestBlock(num *big.Int) *core.Block {
	return &core.Block{
		Header: &core.BlockHeader{
			Time:       uint64(time.Now().Unix()),
			Difficulty: 0,
			Nonce:      0,
			Number:     num,
		},
	}
}

func newP2PNode(privateKey *ecdsa.PrivateKey) *p2p.Node {
	return &p2p.Node{
		PrivateKey: privateKey,
	}
}

func NewTestCoreForSign(privateKey *ecdsa.PrivateKey) *Core {
	return &Core{
		privateKey: privateKey,
		address:    crypto.PubkeyToAddress(privateKey.PublicKey).Bytes(),
	}
}

func newTestCore(p2pNode *p2p.Node) *Core {
	return &Core{
		p2pNode:    p2pNode,
		state:      pbfttypes.StateAcceptRequest,
		privateKey: p2pNode.PrivateKey,
		address:    crypto.PubkeyToAddress(p2pNode.PrivateKey.PublicKey).Bytes(),
	}
}

func TestProcess(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("生成密钥失败：%v", err)
	}

	p2pNode := newP2PNode(privateKey)

	core := newTestCore(p2pNode)

	core.Start()

	// block := newTestBlock()

	// event.Bus.Publish("ConsensusStart", block)

	time.Sleep(2 * time.Second)

	core.Stop()

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
