package crypto

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

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
