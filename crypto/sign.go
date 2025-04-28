package crypto

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/herumi/bls-eth-go-binary/bls"
)

type BLSSig = bls.Sign

// 使用私钥对哈希后的摘要进行签名
func Sign(digestHash []byte, prv *ecdsa.PrivateKey) (sig []byte, err error) {
	return ethcrypto.Sign(digestHash, prv)
}

// 通过哈希和签名得到签名所用的公钥
func SigToPub(hash, sig []byte) (*ecdsa.PublicKey, error) {
	return ethcrypto.SigToPub(hash, sig)
}

// 由公钥得到地址
func PubkeyToAddress(p ecdsa.PublicKey) common.Address {
	return ethcrypto.PubkeyToAddress(p)
}

func BLSSign(msg []byte, blsPrivateKey *BLSPrivKey) (*BLSSig, error) {
	blsSig := blsPrivateKey.SignByte(msg)
	return blsSig, nil
}

// 公钥验证签名；聚合公钥验证聚合签名
func BLSVerify(msg []byte, blsPublicKey *BLSPubKey, blsSig *BLSSig) bool {
	ok := blsSig.VerifyByte(blsPublicKey, msg)
	return ok
}

func AggregateSignatures(sigs []BLSSig) (*BLSSig, error) {
	if len(sigs) == 0 {
		return nil, fmt.Errorf("no signatures to aggregate")
	}
	var agg BLSSig
	agg.Aggregate(sigs)
	return &agg, nil
}

// 验证聚合签名
func VerifyAggregateSignature(pubKeys []BLSPubKey, aggSig *BLSSig, msg []byte) bool {
	if aggSig == nil {
		return false
	}
	aggPubKey, err := AggregatePublicKeys(pubKeys)
	if err != nil {
		return false
	}
	return aggSig.VerifyByte(aggPubKey, msg)
}

func AggregatePublicKeys(pubKeys []BLSPubKey) (*BLSPubKey, error) {
	if len(pubKeys) == 0 {
		return nil, fmt.Errorf("no public keys to aggregate")
	}

	aggPubKey := pubKeys[0] // 直接拿第一个
	for i := 1; i < len(pubKeys); i++ {
		aggPubKey.Add(&pubKeys[i])
	}
	return &aggPubKey, nil
}
