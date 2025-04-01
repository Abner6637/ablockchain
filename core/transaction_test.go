package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTransaction(t *testing.T) {
	stateDB, err := NewStateDB("test_db")
	assert.NoError(t, err)
	defer stateDB.Close()

	from, err := stateDB.NewAccount()
	if err != nil {
		return
	}

	to := "0x456"
	value := uint64(100)
	tx := NewTransaction(from, to, value)

	if tx.From != from.Address {
		t.Errorf("Expected From: %s, got: %s", from.Address, tx.From)
	}
	if tx.To != to {
		t.Errorf("Expected To: %s, got: %s", to, tx.To)
	}
	if tx.Value != value {
		t.Errorf("Expected Value: %d, got: %d", value, tx.Value)
	}
}

func TestEncodeDecodeTransaction(t *testing.T) {
	stateDB, err := NewStateDB("test_db")
	assert.NoError(t, err)
	defer stateDB.Close()

	from, err := stateDB.NewAccount()
	if err != nil {
		return
	}
	to := "0x456"
	value := uint64(100)
	tx := NewTransaction(from, to, value)
	encoded, err := tx.EncodeTx()
	if err != nil {
		t.Fatalf("Failed to encode transaction: %v", err)
	}

	decodedTx, err := DecodeTx(encoded)
	if err != nil {
		t.Fatalf("Failed to decode transaction: %v", err)
	}

	if decodedTx.From != tx.From || decodedTx.To != tx.To || decodedTx.Value != tx.Value {
		t.Errorf("Decoded transaction does not match original")
	}
}

func TestTransactionTimestamp(t *testing.T) {
	stateDB, err := NewStateDB("test_db")
	assert.NoError(t, err)
	defer stateDB.Close()

	from, err := stateDB.NewAccount()
	if err != nil {
		return
	}
	tx := NewTransaction(from, "0x456", 100)
	currentTime := uint64(time.Now().Unix())

	if tx.Time > currentTime || tx.Time < currentTime-1 {
		t.Errorf("Unexpected timestamp: %d", tx.Time)
	}
}
