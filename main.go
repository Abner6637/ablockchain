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

<<<<<<< HEAD
	// 创建区块链并启动一个异步的miner进程
	bc, err := core.NewBlockchain()
	if err != nil {
		return
	}
	bc.StartMiner()
=======
	}
	SHA256Hash := hash1.Hash(data)
	fmt.Printf("SHA-256: %x\n", SHA256Hash)

	hash2, err := crypto.NewKeccak256()
	if err != nil {
		fmt.Printf("G")

	}
	Keccak256Hash := hash2.Hash(data)
	fmt.Printf("Keccak-256: %x\n", Keccak256Hash)
>>>>>>> 43c0f83 (block add merkle)

}
