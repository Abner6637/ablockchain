package crypto

import (
	"crypto/sha256"

	"github.com/ethereum/go-ethereum/crypto"
)

func SHA256Hash(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

func Keccak256Hash(data []byte) []byte {
	return crypto.Keccak256(data) // 通过 go-ethereum 库实现 Keccak-256
}
