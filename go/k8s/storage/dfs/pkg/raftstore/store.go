package raftstore

// Store is the interface that defines the abstract and necessary methods for storage operation.
type Store interface {
	Put(key, val interface{}) (interface{}, error)
	Get(key interface{}) (interface{}, error)
	Del(key interface{}) (interface{}, error)
}
