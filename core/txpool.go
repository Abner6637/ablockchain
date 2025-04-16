package core

import (
	"fmt"
	"sync"
)

// Pending:可以执行的交易，Queue：未来交易（nonce值大于目前账户nonce值）#string为付款账户地址
type TxPool struct {
	Pending []*Transaction
	Queued  []*Transaction
	lock    sync.RWMutex
}

func NewTxPool() *TxPool {
	pool := &TxPool{
		Pending: make([]*Transaction, 0),
		Queued:  make([]*Transaction, 0),
	}
	return pool
}

// 添加交易：Pending 或 Queued
func (tp *TxPool) AddPending(tx *Transaction) {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	tp.Pending = append(tp.Pending, tx)
}

func (tp *TxPool) AddQueued(tx *Transaction) {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	tp.Queued = append(tp.Queued, tx)
}

func (tp *TxPool) RemoveQueued(tx *Transaction) {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	for i := 0; i < len(tp.Queued); i++ {
		if tp.Queued[i] == tx {
			tp.Queued = append(tp.Queued[:i], tp.Queued[i+1:]...)
			i--
			break
		}
	}
}

// 在queued中查询符合addr和nonce的交易
func (tp *TxPool) FindQueuedTxByNonce(from string, nonce uint64) *Transaction {
	tp.lock.RLock()
	defer tp.lock.RUnlock()

	for _, tx := range tp.Queued {
		if tx.From == from && tx.Nonce == nonce {
			return tx
		}
	}
	return nil
}

// 获取Pending中的交易,指定交易个数
func (tp *TxPool) GetPendingTxs(transactionsPerBlock int) []*Transaction {
	tp.lock.RLock()
	defer tp.lock.RUnlock()

	n := transactionsPerBlock
	if n > len(tp.Pending) {
		n = len(tp.Pending)
	}

	txs := make([]*Transaction, n)
	copy(txs, tp.Pending[:n])
	return txs
}

func (tp *TxPool) GetQueuedTxs() []*Transaction {
	tp.lock.RLock()
	defer tp.lock.RUnlock()

	txs := make([]*Transaction, len(tp.Queued))
	copy(txs, tp.Queued)
	return txs
}

func (tp *TxPool) PrintPending() {
	tp.lock.Lock()
	defer tp.lock.Unlock()
	fmt.Println("Pending:\n")
	for _, tx := range tp.Pending {
		tx.PrintTransaction()
	}
}

func (tp *TxPool) PrintQueued() {
	tp.lock.Lock()
	defer tp.lock.Unlock()
	fmt.Println("Queued:\n")
	for _, tx := range tp.Queued {
		tx.PrintTransaction()
	}
}

// 返回待处理交易的数量
func (tp *TxPool) PendingSize() int {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	return len(tp.Pending)
}

func (tp *TxPool) QueuedSize() int {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	return len(tp.Queued)
}

// 返回交易池中from账户的最大nonce值
func (tp *TxPool) GetMaxNonce(from string) uint64 {
	maxNonce := uint64(0)
	for _, tx := range tp.Pending {
		if tx.From == from {
			if tx.Nonce > maxNonce {
				maxNonce = tx.Nonce
			}
		}
	}
	return maxNonce
}

// 交易池中清除已经打包的区块
func (tp *TxPool) ClearPackedTxs(txs []*Transaction) {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	for i := 0; i < len(tp.Pending); i++ {
		for _, tx := range txs {
			if tp.Pending[i] == tx {
				tp.Pending = append(tp.Pending[:i], tp.Pending[i+1:]...)
				i--
				break
			}
		}
	}
	for i := 0; i < len(tp.Queued); i++ {
		for _, tx := range txs {
			if tp.Queued[i] == tx {
				tp.Queued = append(tp.Queued[:i], tp.Queued[i+1:]...)
				i--
				break
			}
		}
	}
}
