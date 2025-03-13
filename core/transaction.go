package core

import "time"

type Transaction struct {
	Time  time.Time
	Hash  []byte
	Nonce uint64
}
