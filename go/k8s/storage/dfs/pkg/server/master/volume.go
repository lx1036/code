package master

import "sync"

// Vol represents a set of meta partitionMap and data partitionMap
type Volume struct {
	ID             uint64
	Name           string
	Owner          string
	s3Endpoint     string
	mpReplicaNum   uint8
	Status         uint8
	threshold      float32
	Capacity       uint64 // GB
	MetaPartitions map[uint64]*MetaPartition
	mpsLock        sync.RWMutex
	mpsCache       []byte
	viewCache      []byte
	bucketdeleted  bool
	createMpMutex  sync.RWMutex
	sync.RWMutex
}
