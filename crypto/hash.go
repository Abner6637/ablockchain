package crypto

import (
	"crypto/sha256"
	"sync"
)

// Hash 接口定义哈希算法的统一接口
type Hash interface {
	Hash(data []byte) []byte
}

type SHA256 struct{}

type Keccak256 struct{}

// GlobalHashAlgorithm 是当前全局哈希算法实例
var GlobalHashAlgorithm Hash
var once sync.Once

// 初始化时设置默认的哈希算法
func init() {
	// 默认使用 NewKeccak256
	once.Do(func() {
		GlobalHashAlgorithm = NewKeccak256()
	})
}

func NewSHA256() *SHA256 {
	return &SHA256{}
}

func NewKeccak256() *Keccak256 {
	return &Keccak256{}
}

// SetGlobalHashAlgorithm 设置全局哈希算法
func SetGlobalHashAlgorithm(hashAlg Hash) {
	GlobalHashAlgorithm = hashAlg
}

func (h *SHA256) Hash(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

func (h *Keccak256) Hash(data []byte) []byte {
	return Keccak256_geth(data)
}
