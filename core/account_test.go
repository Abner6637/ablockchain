package core_test

import (
	"ablockchain/core"
	"os"
	"testing"
)

// 清理测试用的 LevelDB 文件夹
func cleanDB(path string) {
	os.RemoveAll(path)
}

func TestCommitAndTraverseAccounts(t *testing.T) {
	dbPath := "testdb"
	cleanDB(dbPath) // 测试前清除旧数据
	defer cleanDB(dbPath)

	// 初始化 StateDB
	stateDB, err := core.NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create StateDB: %v", err)
	}
	defer stateDB.Close()

	// 创建两个账户
	account1, err := stateDB.NewAccount()
	if err != nil {
		t.Fatalf("Failed to create account1: %v", err)
	}

	account2, err := stateDB.NewAccount()
	if err != nil {
		t.Fatalf("Failed to create account2: %v", err)
	}

	// 修改余额以测试 dirty 状态跟踪
	account1.State.Balance += 50
	stateDB.UpdateAccState(account1.State)

	account2.State.Balance += 100
	stateDB.UpdateAccState(account2.State)

	// Commit 到 Trie/LevelDB
	_, err = stateDB.Commit()
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// 用 TraverseAccounts 遍历 LevelDB，验证账户是否存在
	foundAccounts := make(map[string]bool)
	stateDB.TraverseAccState(func(key string, acc *core.AccState) {
		foundAccounts[acc.Address] = true
		t.Logf("Traversed account: %s | Balance: %d | Nonce: %d", acc.Address, acc.Balance, acc.Nonce)
	})

	// 验证两个账户都被找到
	if !foundAccounts[account1.State.Address] {
		t.Errorf("Account1 not found in traversal")
	}
	if !foundAccounts[account2.State.Address] {
		t.Errorf("Account2 not found in traversal")
	}
}
