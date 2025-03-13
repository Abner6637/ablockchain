package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 测试 BoltStorage 的基本功能
func TestBoltStorage(t *testing.T) {
	// 测试数据库文件
	dbPath := "test.db"
	bucketName := "TestBucket"

	// 创建 BoltStorage 实例
	storage, err := NewBoltDB(dbPath, bucketName)
	assert.NoError(t, err, "创建 BoltStorage 失败")
	defer storage.Close()

	// 清理测试数据库文件
	defer os.Remove(dbPath)

	// 测试数据
	key := "testKey"
	value := map[string]string{"name": "Alice", "email": "alice@example.com"}

	// 测试 Put 方法
	err = storage.Put(key, value)
	assert.NoError(t, err, "Put 方法失败")

	// 测试 Get 方法
	var retrievedValue map[string]string
	err = storage.Get(key, &retrievedValue)
	assert.NoError(t, err, "Get 方法失败")
	assert.Equal(t, value, retrievedValue, "Get 结果不匹配")

	// 测试 Delete 方法
	err = storage.Delete(key)
	assert.NoError(t, err, "Delete 方法失败")

	// 确保删除成功
	err = storage.Get(key, &retrievedValue)
	assert.Error(t, err, "Get 应该返回错误（键已删除）")
}
