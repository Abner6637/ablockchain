package core

type TxPool struct {
	TxQueue []*Transaction
}

func NewTxPool() *TxPool {
	txQueue := make([]*Transaction, 0)
	return &TxPool{TxQueue: txQueue}
}
