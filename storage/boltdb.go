package storage

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"go.etcd.io/bbolt"
)

// BoltStorage 基于 BoltDB 实现 Storage 接口
type BoltStorage struct {
	db     *bbolt.DB
	bucket []byte
}

// NewBoltStorage 创建 BoltDB 实例
func NewBoltDB(dbPath string, bucketName string) (*BoltStorage, error) {
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, err
	}

	// 确保 Bucket 存在
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	return &BoltStorage{db: db, bucket: []byte(bucketName)}, nil
}

// Put 存储键值对
func (b *BoltStorage) Put(key string, value interface{}) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(b.bucket)
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}

		// 序列化数据
		var buffer bytes.Buffer
		enc := gob.NewEncoder(&buffer)
		if err := enc.Encode(value); err != nil {
			return err
		}

		return bucket.Put([]byte(key), buffer.Bytes())
	})
}

// Get 读取键值对
func (b *BoltStorage) Get(key string, value interface{}) error {
	return b.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(b.bucket)
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}

		data := bucket.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("key not found")
		}

		// 反序列化数据
		buffer := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buffer)
		return dec.Decode(value)
	})
}

// Delete 删除键
func (b *BoltStorage) Delete(key string) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(b.bucket)
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}
		return bucket.Delete([]byte(key))
	})
}

// Close 关闭数据库
func (b *BoltStorage) Close() error {
	return b.db.Close()
}
