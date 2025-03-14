package core

import "sync"

type TxPool struct {
	TxQueue []*Transaction
	lock    sync.RWMutex
}

func NewTxPool() *TxPool {
	txQueue := make([]*Transaction, 0)
	return &TxPool{TxQueue: txQueue}
}

// 添加交易
func (tp *TxPool) AddTx(tx *Transaction) {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	tp.TxQueue = append(tp.TxQueue, tx)
}

// 获取交易
func (tp *TxPool) GetTxs() []*Transaction {
	tp.lock.RLock()
	defer tp.lock.RUnlock()

	txs := make([]*Transaction, len(tp.TxQueue))
	copy(txs, tp.TxQueue)
	return txs
}

// 返回待处理交易的数量
func (tp *TxPool) PendingSize() int {
	return len(tp.TxQueue)
}
