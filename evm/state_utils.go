package evm

import (
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
)

// StateDBWithAlloc creates a statedb and populates with the given alloc.
func StateDBWithAlloc(alloc types.GenesisAlloc) *state.StateDB {
	statedb := NewEmptyStateDB()
	for addr, acc := range alloc {
		statedb.CreateAccount(addr)
		statedb.SetCode(addr, acc.Code)
		statedb.SetNonce(addr, acc.Nonce, 0)
		if acc.Balance != nil {
			statedb.SetBalance(addr, uint256.MustFromBig(acc.Balance), tracing.BalanceChangeUnspecified)
		}
	}
	return statedb
}

// StateDBWithAlloc creates an empty statedb, or panics on error
func NewEmptyStateDB() *state.StateDB {
	statedb, err := state.New(types.EmptyRootHash, NewMemStateDB())
	if err != nil {
		panic(err)
	}
	return statedb
}

func NewMemStateDB() state.Database {
	return state.NewDatabase(triedb.NewDatabase(rawdb.NewMemoryDatabase(), nil), nil)
}
