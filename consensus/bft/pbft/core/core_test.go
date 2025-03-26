package pbftcore

import (
	"ablockchain/consensus/bft"
	pbfttypes "ablockchain/consensus/bft/pbft/types"
	"ablockchain/crypto"
	"bytes"
	"crypto/ecdsa"
	"log"
	"math/big"
	"testing"
)

func SignNewCore(privateKey *ecdsa.PrivateKey) *Core {
	return &Core{
		privateKey: privateKey,
		address:    crypto.PubkeyToAddress(privateKey.PublicKey).Bytes(),
	}
}

func TestSign(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	log.Printf("privateKey: %+v", privateKey)

	if err != nil {
		t.Fatalf("生成密钥失败：%v", err)
	}

	core := SignNewCore(privateKey)

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

	// 比较签名地址和消息中的地址参数（即发送消息的地址）是否一致
	if !bytes.Equal(signerAddress.Bytes(), msg.Address) {
		t.Fatalf("验证失败")
	} else {
		log.Printf("地址相同，验证成功")
	}
}
