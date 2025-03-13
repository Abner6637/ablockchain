package core

import "ablockchain/storage"

type Blockchain struct {
	db     *storage.LevelDB
	TxPool *TxPool
}
