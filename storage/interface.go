package storage

// Storage 通用数据库接口
type Storage interface {
	Put(key string, value interface{}) error
	Get(key string, value interface{}) error
	Delete(key string) error
	Close() error
}
