package raftstore

// INFO: memory 作为后端 meta 数据后端存储引擎，后续改成 boltDB
type MemoryStore struct {
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}
