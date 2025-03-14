```go
package main

import (
	"ablockchain/crypto"
	"fmt"
)

func main() {
	data := []byte("Hello, Ethereum!")

	hash1, err := crypto.NewSHA256()
	if err != nil {
		fmt.Printf("G")

	}
	SHA256Hash := hash1.Hash(data)
	fmt.Printf("SHA-256: %x\n", SHA256Hash)

	hash2, err := crypto.NewKeccak256()
	if err != nil {
		fmt.Printf("G")

	}
	Keccak256Hash := hash2.Hash(data)
	fmt.Printf("Keccak-256: %x\n", Keccak256Hash)

}
```