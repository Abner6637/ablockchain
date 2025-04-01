package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// 测试创建新账户
func TestNewAccount(t *testing.T) {
	stateDB, err := NewStateDB("test_db")
	assert.NoError(t, err)
	defer stateDB.Close()

	account, err := stateDB.NewAccount()
	assert.NoError(t, err)
	assert.NotNil(t, account)
	assert.NotEmpty(t, account.Address)
	assert.NotEmpty(t, account.PublicKey)
	assert.NotEmpty(t, account.PrivateKey)
	assert.Equal(t, uint64(0), account.Balance)

	t.Logf("新账户创建成功: 地址 %s", account.Address)
}

// 测试更新账户余额
func TestUpdateAccount(t *testing.T) {
	stateDB, err := NewStateDB("test_db")
	assert.NoError(t, err)
	defer stateDB.Close()

	account, err := stateDB.NewAccount()
	assert.NoError(t, err)

	stateDB.PrintAccounts()

	// 修改余额
	account.Balance = 1000
	err = stateDB.UpdateAccount(account)
	assert.NoError(t, err)

	// 查询账户，检查余额是否更新
	updatedAccount, exists := stateDB.GetAccount(account.Address)
	assert.True(t, exists)
	assert.Equal(t, uint64(1000), updatedAccount.Balance)
	stateDB.PrintAccounts()
}

// 测试查询账户
func TestGetAccount(t *testing.T) {
	stateDB, err := NewStateDB("test_db")
	assert.NoError(t, err)
	defer stateDB.Close()

	account, err := stateDB.NewAccount()
	assert.NoError(t, err)

	// 查询账户
	retrievedAccount, exists := stateDB.GetAccount(account.Address)
	assert.True(t, exists)
	assert.Equal(t, account.Address, retrievedAccount.Address)
	assert.Equal(t, account.Balance, retrievedAccount.Balance)
}

// 测试查询不存在的账户
func TestGetNonExistentAccount(t *testing.T) {
	stateDB, err := NewStateDB("test_db")
	assert.NoError(t, err)
	defer stateDB.Close()

	account, exists := stateDB.GetAccount("0x123456789")
	assert.False(t, exists)
	assert.Nil(t, account)

	t.Log("查询不存在的账户，返回 nil")
}

func TestVerifySignature(t *testing.T) {
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

	// 签名交易
	signature, err := from.SignTx(tx)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	// 验证签名
	valid, err := tx.VerifySignature(signature.Sign)
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}

	if !valid {
		t.Errorf("Signature verification failed")
	}
}
