package checksum

import (
	"hash/fnv"

	hashutil "k8s.io/kubernetes/pkg/util/hash"
)

// Checksum is the data to be stored as checkpoint
type Checksum uint64

// New returns the Checksum of checkpoint data
func New(data interface{}) Checksum {
	return Checksum(getChecksum(data))
}

// Get returns calculated checksum of checkpoint data
func getChecksum(data interface{}) uint64 {
	hash := fnv.New32a()
	hashutil.DeepHashObject(hash, data)
	return uint64(hash.Sum32())
}
