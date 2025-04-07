package storage

import (
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
)

// 创建临时数据库路径
const testDBPath = "./test_leveldb"

// 测试数据库初始化
func TestNewLevelDB(t *testing.T) {
	db, err := NewLevelDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to initialize LevelDB: %v", err)
	}
	defer db.Close()
}

// 测试存储和读取数据
func TestLevelDB_PutAndGet(t *testing.T) {
	db, err := NewLevelDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to open LevelDB: %v", err)
	}
	defer db.Close()

	// 存储数据
	err = db.Put("testKey", "testValue")
	if err != nil {
		t.Fatalf("Failed to put data: %v", err)
	}

	// 读取数据
	var value string
	err = db.Get("testKey", &value)
	if err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}

	// 验证数据是否正确
	if value != "testValue" {
		t.Errorf("Expected 'testValue', got '%s'", value)
	}
}

// 测试存储整数
func TestLevelDB_PutAndGetInt(t *testing.T) {
	db, err := NewLevelDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to open LevelDB: %v", err)
	}
	defer db.Close()

	// 存储 int 类型
	err = db.Put("testInt", 42)
	if err != nil {
		t.Fatalf("Failed to put int: %v", err)
	}

	// 读取 int 类型
	var value int
	err = db.Get("testInt", &value)
	if err != nil {
		t.Fatalf("Failed to get int: %v", err)
	}

	// 验证数据是否正确
	if value != 42 {
		t.Errorf("Expected 42, got %d", value)
	}
}

// 测试删除数据
func TestLevelDB_Delete(t *testing.T) {
	db, err := NewLevelDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to open LevelDB: %v", err)
	}
	defer db.Close()

	// 存储数据
	err = db.Put("deleteKey", "toBeDeleted")
	if err != nil {
		t.Fatalf("Failed to put data: %v", err)
	}

	// 删除数据
	err = db.Delete("deleteKey")
	if err != nil {
		t.Fatalf("Failed to delete data: %v", err)
	}

	// 尝试读取已删除的数据
	var value string
	err = db.Get("deleteKey", &value)
	if err == nil {
		t.Errorf("Expected error when getting deleted key, but got value: %s", value)
	}
}

// 清理测试数据库
func TestMain(m *testing.M) {
	// 运行测试
	code := m.Run()

	// 清理测试数据库
	os.RemoveAll(testDBPath)

	// 退出测试
	os.Exit(code)
}

// 假设你有一个结构体用于测试
type testAccount struct {
	Address string
	Balance uint64
}

func TestLevelDB_ForEach(t *testing.T) {
	// 创建临时数据库路径
	dbPath := t.TempDir()

	defer os.RemoveAll(dbPath) // 测试完成后清理

	db, err := NewLevelDB(dbPath)
	if err != nil {
		t.Fatalf("打开 LevelDB 失败: %v", err)
	}
	defer db.Close()

	// 准备测试数据
	accounts := []testAccount{
		{"0x123", 100},
		{"0x456", 200},
	}

	// 存入测试数据
	for _, acc := range accounts {
		data, err := rlp.EncodeToBytes(acc)
		if err != nil {
			t.Fatalf("编码失败: %v", err)
		}
		if err := db.Put(acc.Address, data); err != nil {
			t.Fatalf("存入失败: %v", err)
		}
	}

	// 遍历验证
	found := make(map[string]bool)
	db.ForEach(func(key string, value []byte) {
		var acc testAccount
		if err := rlp.DecodeBytes(value, &acc); err == nil {
			found[acc.Address] = true
		}
	})

	// 验证是否都找到了
	for _, acc := range accounts {
		if !found[acc.Address] {
			t.Errorf("未找到账户: %s", acc.Address)
		}
	}
}
