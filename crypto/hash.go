package crypto

import (
	"crypto/sha256"

	"github.com/ethereum/go-ethereum/crypto"
)

type Hash interface {
	Hash(data []byte) []byte
}

type SHA256 struct{}

type Keccak256 struct{}

func NewSHA256() *SHA256 {
	hash := new(SHA256)
	return hash
}

func NewKeccak256() *Keccak256 {
	hash := new(Keccak256)
	return hash
}

func (h *SHA256) Hash(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

func (h *Keccak256) Hash(data []byte) []byte {
	return crypto.Keccak256(data) // 通过 go-ethereum 库实现 Keccak-256
}
