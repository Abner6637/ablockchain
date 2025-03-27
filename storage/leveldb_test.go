package storage

import (
	"os"
	"testing"
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
	db.PrintCount()
	db.GetLatest()
	db.GetAll()
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
