package core

import (
	"sync"
)

// Pending:可以执行的交易，Queue：未来交易（nonce值大于目前账户nonce值）#string为付款账户地址
type TxPool struct {
	Pending      map[string]*TxList
	Queued       map[string]*TxList
	CurrentNonce map[string]uint64 //暂存nonce
	lock         sync.RWMutex
}

func NewTxPool() *TxPool {
	pool := &TxPool{
		Pending:      make(map[string]*TxList),
		Queued:       make(map[string]*TxList),
		CurrentNonce: make(map[string]uint64),
	}
	return pool
}

// nonce => tx
type TxList struct {
	items map[uint64]*Transaction
}

func NewTxList(tx *Transaction, nonce uint64) *TxList {
	return &TxList{
		items: map[uint64]*Transaction{
			nonce: tx,
		},
	}
}

// 添加交易：Pending或Queued
func (tp *TxPool) AddTx(tx *Transaction) {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	txNonce := tx.Nonce
	currentnonce, ok := tp.CurrentNonce[tx.From]
	if !ok { //该账户的第一笔交易
		tp.CurrentNonce[tx.From] = 1
		txlist := NewTxList(tx, tx.Nonce)
		tp.Pending[tx.From] = txlist

	} else {
		if txNonce == currentnonce {
			// 可以执行，加入 pending 区
			txlist := NewTxList(tx, tx.Nonce)
			tp.Pending[tx.From] = txlist
		} else if txNonce > currentnonce {
			// 非连续的交易，加入 queued（挂起）
			txlist := NewTxList(tx, tx.Nonce)
			tp.Queued[tx.From] = txlist
			//tp.CurrentNonce[tx.From] += 1
		}
		// nonce 太小了，直接拒绝
	}

}

// 定期将queue中交易加入pending //待测试
func (tp *TxPool) PromoteExecutables() {
	for addr, txlist := range tp.Queued {
		currentNonce := tp.CurrentNonce[addr]
		for {
			tx := txlist.items[currentNonce]
			if tx == nil {
				break // 下一个 nonce 的交易还没到
			}
			// 当前 nonce 的交易存在，移动到 pending
			txlist := NewTxList(tx, tx.Nonce)
			tp.Pending[addr] = txlist
			delete(txlist.items, currentNonce)
			currentNonce++
		}
	}
}

// 获取Pending中的交易
func (tp *TxPool) GetPendingTxs() []*Transaction {
	tp.lock.RLock()
	defer tp.lock.RUnlock()

	var txs []*Transaction

	for _, txList := range tp.Pending {
		for _, tx := range txList.items {
			txs = append(txs, tx)
		}
	}

	return txs
}

func (tp *TxPool) GetAllTxs() []*Transaction {
	tp.lock.RLock()
	defer tp.lock.RUnlock()

	var txs []*Transaction

	// Pending 区交易
	for _, txList := range tp.Pending {
		for _, tx := range txList.items {
			txs = append(txs, tx)
		}
	}

	// Queued 区交易
	for _, txList := range tp.Queued {
		for _, tx := range txList.items {
			txs = append(txs, tx)
		}
	}

	return txs
}

// 返回待处理交易的数量
func (tp *TxPool) PendingSize() int {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	return len(tp.Pending)
}

// 交易池中清除已经打包的区块
func (tp *TxPool) ClearPackedTxs(txs []*Transaction) {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	for _, tx := range txs {

		nonce := tx.Nonce

		// 从 Pending 中删除
		if txList, ok := tp.Pending[tx.From]; ok {
			delete(txList.items, nonce)
			if len(txList.items) == 0 {
				delete(tp.Pending, tx.From)
			}
		}

		// 从 Queued 中删除
		if txList, ok := tp.Queued[tx.From]; ok {
			delete(txList.items, nonce)
			if len(txList.items) == 0 {
				delete(tp.Queued, tx.From)
			}
		}
	}
}
