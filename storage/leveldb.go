package storage

import (
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/syndtr/goleveldb/leveldb"
)

// LevelDBStorage 实现 Storage 接口
type LevelDB struct {
	count uint64 // 计数器
	db    *leveldb.DB
}

// NewLevelDBStorage 创建 LevelDB 实例
func NewLevelDB(path string) (*LevelDB, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &LevelDB{db: db, count: 0}, nil
}

// Put 存储数据
func (s *LevelDB) Put(key string, value interface{}) error {
	data, err := rlp.EncodeToBytes(value) // RLP 编码
	if err != nil {
		return err
	}
	s.count++
	return s.db.Put([]byte(key), data, nil)
}

// Get 读取数据
func (s *LevelDB) Get(key string, value interface{}) error {
	data, err := s.db.Get([]byte(key), nil)
	if err != nil {
		return err
	}
	// RLP 解码
	return rlp.DecodeBytes(data, value)
}

// Delete 删除键
func (s *LevelDB) Delete(key string) error {
	s.count--
	return s.db.Delete([]byte(key), nil)
}

// 打印键值对总数
func (s *LevelDB) PrintCount() {
	fmt.Printf("Total: %d\n", s.count)
}

// 返回rlp编码，手动进行解析（block或transaction等)
func (s *LevelDB) GetLatest() (string, []byte, error) {
	iter := s.db.NewIterator(nil, nil)
	defer iter.Release()

	if !iter.Last() {
		return "", nil, leveldb.ErrNotFound
	}

	key := string(iter.Key())
	data := iter.Value()

	return key, data, nil
}

// 获取所有键值对（返回列表）
func (s *LevelDB) GetAll() ([]struct {
	Key   string
	Value []byte
}, error) {
	iter := s.db.NewIterator(nil, nil)
	defer iter.Release()

	var result []struct {
		Key   string
		Value []byte
	}

	for iter.Next() {
		result = append(result, struct {
			Key   string
			Value []byte
		}{
			Key:   string(iter.Key()),
			Value: iter.Value(),
		})
	}

	if len(result) == 0 {
		return nil, leveldb.ErrNotFound
	}
	return result, nil
}

// Close 关闭数据库
func (s *LevelDB) Close() error {
	return s.db.Close()
}
