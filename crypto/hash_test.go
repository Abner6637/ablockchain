package crypto

import (
	"crypto/sha256"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

func TestSHA256Hash(t *testing.T) {
	// 输入数据
	data := []byte("Hello, Ethereum!")

	// 预期的哈希值 (这里是你计算过的 SHA256 哈希值)
	expectedHash := sha256.Sum256(data)

	// 调用 SHA256Hash 函数
	result := SHA256Hash(data)

	// 比较结果
	if !compareByteSlices(result, expectedHash[:]) {
		t.Errorf("SHA256Hash failed, expected %x, got %x", expectedHash, result)
	}
}

func TestKeccak256Hash(t *testing.T) {
	// 输入数据
	data := []byte("Hello, Ethereum!")

	// 预期的 Keccak-256 哈希值（可以通过在线工具或代码预先计算得到）
	expectedHash := crypto.Keccak256(data)

	// 调用 Keccak256Hash 函数
	result := Keccak256Hash(data)

	// 比较结果
	if !compareByteSlices(result, expectedHash) {
		t.Errorf("Keccak256Hash failed, expected %x, got %x", expectedHash, result)
	}
}

// compareByteSlices 用于比较两个字节切片是否相等
func compareByteSlices(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
