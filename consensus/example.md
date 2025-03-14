```go
package main

import (
	"ablockchain/consensus"
	"ablockchain/consensus/poa"
	"ablockchain/consensus/pow"
)

func main() {
	var con consensus.Consensus

	// 直接赋值并调用方法
	con = pow.NewPoW()
	con.Start()
	con.Stop()

	con = poa.NewPoA()
	con.Start()
	con.Stop()
}

```
