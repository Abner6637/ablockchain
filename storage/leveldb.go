package storage

import (
	"encoding/json"

	"github.com/syndtr/goleveldb/leveldb"
)

// LevelDBStorage 实现 Storage 接口
type LevelDB struct {
	db *leveldb.DB
}

// NewLevelDBStorage 创建 LevelDB 实例
func NewLevelDB(path string) (*LevelDB, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &LevelDB{db: db}, nil
}

// Put 存储数据（支持任意类型）
func (s *LevelDB) Put(key string, value interface{}) error {
	data, err := json.Marshal(value) // 序列化为 JSON
	if err != nil {
		return err
	}
	return s.db.Put([]byte(key), data, nil)
}

// Get 读取数据（支持任意类型）
func (s *LevelDB) Get(key string, value interface{}) error {
	data, err := s.db.Get([]byte(key), nil)
	if err != nil {
		return err
	}
	// 反序列化为传入的类型
	return json.Unmarshal(data, value)
}

// Delete 删除键
func (s *LevelDB) Delete(key string) error {
	return s.db.Delete([]byte(key), nil)
}

// Close 关闭数据库
func (s *LevelDB) Close() error {
	return s.db.Close()
}
