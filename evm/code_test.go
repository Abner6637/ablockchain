package evm

import (
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb/leveldb"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
)

func createTestStateDB(t *testing.T) (*state.StateDB, func()) {
	dir := t.TempDir()
	ldb, err := leveldb.New(dir, 256, 512, "", false)
	if err != nil {
		t.Fatalf("Failed to create leveldb: %v", err)
	}

	db := rawdb.NewDatabase(ldb)
	triedb := triedb.NewDatabase(db, nil)

	statedb, err := state.New(types.EmptyRootHash, state.NewDatabase(triedb, nil))
	if err != nil {
		panic(err)
	}

	cleanup := func() {
		ldb.Close()
		os.RemoveAll(dir)
	}

	return statedb, cleanup
}

func TestAccountCreation(t *testing.T) {
	statedb, cleanup := createTestStateDB(t)
	defer cleanup()

	addr := common.HexToAddress("0x01")

	// Test account doesn't exist initially
	if statedb.Exist(addr) {
		t.Error("Account should not exist initially")
	}

	// Create account
	statedb.CreateAccount(addr)
	if !statedb.Exist(addr) {
		t.Error("Account should exist after creation")
	}
	if !statedb.Empty(addr) {
		t.Error("New account should be empty")
	}
}

func TestAccountBalance(t *testing.T) {
	statedb, cleanup := createTestStateDB(t)
	defer cleanup()

	addr := common.HexToAddress("0x02")
	amount := uint256.NewInt(100)

	// Test initial balance is zero
	if statedb.GetBalance(addr).Cmp(uint256.NewInt(0)) != 0 {
		t.Error("Initial balance should be zero")
	}

	// Add balance
	statedb.AddBalance(addr, amount, tracing.BalanceChangeUnspecified)
	if statedb.GetBalance(addr).Cmp(amount) != 0 {
		t.Errorf("Balance should be %v, got %v", amount, statedb.GetBalance(addr))
	}

	// Sub balance
	statedb.SubBalance(addr, uint256.NewInt(50), tracing.BalanceChangeUnspecified)
	if statedb.GetBalance(addr).Cmp(uint256.NewInt(50)) != 0 {
		t.Errorf("Balance should be 50, got %v", statedb.GetBalance(addr))
	}

	// Set balance directly
	statedb.SetBalance(addr, uint256.NewInt(200), tracing.BalanceChangeUnspecified)
	if statedb.GetBalance(addr).Cmp(uint256.NewInt(200)) != 0 {
		t.Errorf("Balance should be 200, got %v", statedb.GetBalance(addr))
	}
}

func TestAccountNonce(t *testing.T) {
	statedb, cleanup := createTestStateDB(t)
	defer cleanup()

	addr := common.HexToAddress("0x03")

	// Test initial nonce is zero
	if statedb.GetNonce(addr) != 0 {
		t.Error("Initial nonce should be zero")
	}

	// Set nonce
	statedb.SetNonce(addr, 10, tracing.NonceChangeUnspecified)
	if statedb.GetNonce(addr) != 10 {
		t.Errorf("Nonce should be 10, got %d", statedb.GetNonce(addr))
	}
}

func TestAccountCode(t *testing.T) {
	statedb, cleanup := createTestStateDB(t)
	defer cleanup()

	addr := common.HexToAddress("0x04")
	code := []byte{0x01, 0x02, 0x03}

	// Test initial code is empty
	if statedb.GetCode(addr) != nil {
		t.Error("Initial code should be nil")
	}
	if statedb.GetCodeSize(addr) != 0 {
		t.Error("Initial code size should be zero")
	}
	if statedb.GetCodeHash(addr) != (common.Hash{}) {
		t.Error("Initial code hash should be empty hash")
	}

	// Set code
	statedb.SetCode(addr, code)
	if len(statedb.GetCode(addr)) != len(code) {
		t.Errorf("Code length mismatch, want %d, got %d", len(code), len(statedb.GetCode(addr)))
	}
	if statedb.GetCodeSize(addr) != len(code) {
		t.Errorf("Code size mismatch, want %d, got %d", len(code), statedb.GetCodeSize(addr))
	}
	if statedb.GetCodeHash(addr) == (common.Hash{}) {
		t.Error("Code hash should not be empty after setting code")
	}
}

func TestAccountSelfDestruct(t *testing.T) {
	statedb, cleanup := createTestStateDB(t)
	defer cleanup()

	addr := common.HexToAddress("0x05")
	amount := uint256.NewInt(100)

	// Set up account with balance
	statedb.CreateAccount(addr)
	statedb.AddBalance(addr, amount, tracing.BalanceChangeUnspecified)

	// Test self-destruct
	prevBalance := statedb.SelfDestruct(addr)
	if prevBalance.Cmp(amount) != 0 {
		t.Errorf("SelfDestruct should return previous balance %v, got %v", amount, prevBalance)
	}
	if !statedb.HasSelfDestructed(addr) {
		t.Error("Account should be marked as self-destructed")
	}
	if statedb.GetBalance(addr).Cmp(uint256.NewInt(0)) != 0 {
		t.Error("Balance should be zero after self-destruct")
	}

	// Account should still exist until commit
	if !statedb.Exist(addr) {
		t.Error("Account should still exist in state until commit")
	}
}

func TestAccountEmpty(t *testing.T) {
	statedb, cleanup := createTestStateDB(t)
	defer cleanup()

	addr := common.HexToAddress("0x06")

	// New account should be empty
	statedb.CreateAccount(addr)
	if !statedb.Empty(addr) {
		t.Error("New account should be empty")
	}

	// Setting balance makes it non-empty
	statedb.AddBalance(addr, uint256.NewInt(1), tracing.BalanceChangeUnspecified)
	if statedb.Empty(addr) {
		t.Error("Account with balance should not be empty")
	}

	// Reset balance
	statedb.SetBalance(addr, uint256.NewInt(0), tracing.BalanceChangeUnspecified)

	// Setting nonce makes it non-empty
	statedb.SetNonce(addr, 1, tracing.NonceChangeUnspecified)
	if statedb.Empty(addr) {
		t.Error("Account with nonce should not be empty")
	}

	// Reset nonce
	statedb.SetNonce(addr, 0, tracing.NonceChangeUnspecified)

	// Setting code makes it non-empty
	statedb.SetCode(addr, []byte{0x01})
	if statedb.Empty(addr) {
		t.Error("Account with code should not be empty")
	}
}

func TestAccountCommitAndReload(t *testing.T) {
	statedb, cleanup := createTestStateDB(t)
	defer cleanup()

	addr := common.HexToAddress("0x07")
	amount := uint256.NewInt(100)
	nonce := uint64(5)
	code := []byte{0x01, 0x02, 0x03}

	// Set up account
	statedb.CreateAccount(addr)
	statedb.AddBalance(addr, amount, tracing.BalanceChangeUnspecified)
	statedb.SetNonce(addr, nonce, tracing.NonceChangeUnspecified)
	statedb.SetCode(addr, code)

	// Commit state
	root, err := statedb.Commit(0, false, false)
	if err != nil {
		t.Fatalf("Failed to commit state: %v", err)
	}

	// Create new state with the committed root
	newStatedb, err := state.New(root, statedb.Database())
	if err != nil {
		t.Fatalf("Failed to create new state: %v", err)
	}

	// Verify account data persisted
	if !newStatedb.Exist(addr) {
		t.Error("Account should exist after commit")
	}
	if newStatedb.GetBalance(addr).Cmp(amount) != 0 {
		t.Errorf("Balance mismatch after commit, want %v, got %v", amount, newStatedb.GetBalance(addr))
	}
	if newStatedb.GetNonce(addr) != nonce {
		t.Errorf("Nonce mismatch after commit, want %d, got %d", nonce, newStatedb.GetNonce(addr))
	}
	if len(newStatedb.GetCode(addr)) != len(code) {
		t.Errorf("Code length mismatch after commit, want %d, got %d", len(code), len(newStatedb.GetCode(addr)))
	}
}

func TestAccountCopy(t *testing.T) {
	statedb, cleanup := createTestStateDB(t)
	defer cleanup()

	addr := common.HexToAddress("0x08")
	amount := uint256.NewInt(100)

	// Set up account
	statedb.CreateAccount(addr)
	statedb.AddBalance(addr, amount, tracing.BalanceChangeUnspecified)

	// Copy state
	copied := statedb.Copy()

	// Modify original
	statedb.AddBalance(addr, amount, tracing.BalanceChangeUnspecified)

	// Verify copied state is unchanged
	if copied.GetBalance(addr).Cmp(amount) != 0 {
		t.Errorf("Copied balance should be %v, got %v", amount, copied.GetBalance(addr))
	}

	// Verify original was modified
	if statedb.GetBalance(addr).Cmp(uint256.NewInt(200)) != 0 {
		t.Errorf("Original balance should be 200, got %v", statedb.GetBalance(addr))
	}
}

func TestAccountSnapshotAndRevert(t *testing.T) {
	statedb, cleanup := createTestStateDB(t)
	defer cleanup()

	addr := common.HexToAddress("0x09")
	amount := uint256.NewInt(100)

	// Take initial snapshot
	snap1 := statedb.Snapshot()

	// Create account and set balance
	statedb.CreateAccount(addr)
	statedb.AddBalance(addr, amount, tracing.BalanceChangeUnspecified)

	// Take another snapshot
	snap2 := statedb.Snapshot()

	// Modify balance
	statedb.AddBalance(addr, amount, tracing.BalanceChangeUnspecified)

	// Revert to snap2
	statedb.RevertToSnapshot(snap2)
	if statedb.GetBalance(addr).Cmp(amount) != 0 {
		t.Errorf("Balance after revert should be %v, got %v", amount, statedb.GetBalance(addr))
	}

	// Revert to snap1
	statedb.RevertToSnapshot(snap1)
	if statedb.Exist(addr) {
		t.Error("Account should not exist after revert to initial snapshot")
	}
}
