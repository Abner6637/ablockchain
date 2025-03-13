```go
	data := []byte("Hello, Ethereum!")

	Keccak256Hash := crypto.Keccak256Hash(data)
	fmt.Printf("SHA-256: %x\n", sha256Hash)

    sha256Hash := crypto.SHA256Hash(data)
	fmt.Printf("SHA-256: %x\n", sha256Hash)
```