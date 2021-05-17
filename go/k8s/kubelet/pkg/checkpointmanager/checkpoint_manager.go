package checkpointmanager

import (
	"sync"

	utilstore "k8s-lx1036/k8s/kubelet/pkg/util/store"

	utilfs "k8s.io/kubernetes/pkg/util/filesystem"
)

// Checkpoint provides the process checkpoint data
type Checkpoint interface {
	MarshalCheckpoint() ([]byte, error)
	UnmarshalCheckpoint(blob []byte) error
	VerifyChecksum() error
}

// CheckpointManager provides the interface to manage checkpoint
type CheckpointManager interface {
	// CreateCheckpoint persists checkpoint in CheckpointStore. checkpointKey is the key for utilstore to locate checkpoint.
	// For file backed utilstore, checkpointKey is the file name to write the checkpoint data.
	CreateCheckpoint(checkpointKey string, checkpoint Checkpoint) error
	// GetCheckpoint retrieves checkpoint from CheckpointStore.
	GetCheckpoint(checkpointKey string, checkpoint Checkpoint) error
	// WARNING: RemoveCheckpoint will not return error if checkpoint does not exist.
	RemoveCheckpoint(checkpointKey string) error
	// ListCheckpoint returns the list of existing checkpoints.
	ListCheckpoints() ([]string, error)
}

// impl is an implementation of CheckpointManager. It persists checkpoint in CheckpointStore
type impl struct {
	path  string
	store utilstore.Store
	mutex sync.Mutex
}

func (manager *impl) CreateCheckpoint(checkpointKey string, checkpoint Checkpoint) error {
	panic("implement me")
}

func (manager *impl) GetCheckpoint(checkpointKey string, checkpoint Checkpoint) error {
	panic("implement me")
}

func (manager *impl) RemoveCheckpoint(checkpointKey string) error {
	panic("implement me")
}

func (manager *impl) ListCheckpoints() ([]string, error) {
	panic("implement me")
}

// NewCheckpointManager returns a new instance of a checkpoint manager
func NewCheckpointManager(checkpointDir string) (CheckpointManager, error) {
	fstore, err := utilstore.NewFileStore(checkpointDir, utilfs.DefaultFs{})
	if err != nil {
		return nil, err
	}

	return &impl{path: checkpointDir, store: fstore}, nil
}
