package core

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

// 测试辅助函数 - 创建测试区块链
func setupTestBlockchain(t *testing.T) *Blockchain {
	bc, err := NewTestBlockchain("testdb")
	assert.NoError(t, err)
	t.Cleanup(func() {
		// 测试结束后清理数据库
		bc.DB.Close()
	})
	return bc
}

// 测试合约创建
func TestContractCreation(t *testing.T) {
	bc := setupTestBlockchain(t)

	// 简单合约字节码 - 空合约
	code := []byte{0x60, 0x60, 0x60, 0x40, 0x52} // PUSH1 0x60 PUSH1 0x60 PUSH1 0x40 MSTORE

	// 创建合约
	account, err := bc.NewContract(1000, code)
	assert.NoError(t, err)
	assert.NotEmpty(t, account.State.Address)

	// 验证合约状态
	addr := common.HexToAddress(account.State.Address)
	assert.True(t, bc.StateDB.Exist(addr))
	assert.Equal(t, uint64(1), bc.StateDB.GetNonce(addr))
	assert.Equal(t, code, bc.StateDB.GetCode(addr))
}

// 测试简单合约调用 - 返回固定值
func TestSimpleContractCall(t *testing.T) {
	bc := setupTestBlockchain(t)

	// 部署一个简单合约 (返回42)
	codeHex := "6080604052348015600f57600080fd5b506004361060285760003560e01c80636d4ce63c14602d575b600080fd5b60336047565b604051603e91906067565b60405180910390f35b6000602a905090565b6000819050919050565b6061816050565b82525050565b6000602082019050607a6000830184605a565b9291505056fea2646970667358221220857e54f745071af59d359e42dfacba868b7c6695cb"
	code, err := hex.DecodeString(codeHex)
	assert.NoError(t, err, "Failed to decode bytecode")
	contract, err := bc.NewContract(1000, code)
	acc, err := bc.NewAccount()
	assert.NoError(t, err)

	// 合约ABI
	abiJSON := `[{"inputs":[],"name":"get","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]`

	// 调用get()函数
	result, gasUsed, err := bc.RunCode(
		acc.State.Address, // 任意发送者地址
		contract.State.Address,
		abiJSON,
		1000000,
		"get",
	)
	assert.NoError(t, err)
	t.Logf("Gas used: %d", gasUsed)

	// 验证返回值
	assert.Equal(t, big.NewInt(42), result)
}
