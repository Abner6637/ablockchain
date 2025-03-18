```go
package main

import (
	"ablockchain/crypto" // 请替换为你实际使用的包路径
	"fmt"
)

func main() {
	// 使用默认的 SHA256 哈希算法
	data := []byte("hello, world")
	hash := crypto.GlobalHashAlgorithm.Hash(data)
	fmt.Printf("SHA256 Hash: %x\n", hash)

	// 切换为 Keccak256 哈希算法
	crypto.SetGlobalHashAlgorithm(crypto.NewKeccak256())
	hash = crypto.GlobalHashAlgorithm.Hash(data)
	fmt.Printf("Keccak256 Hash: %x\n", hash)
}

```