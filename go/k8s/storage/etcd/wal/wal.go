package wal

import (
	"go.etcd.io/etcd/pkg/fileutil"
	"go.etcd.io/etcd/raft/raftpb"
	"go.etcd.io/etcd/wal/walpb"
	"hash/crc32"
	"os"
	"sync"
	"go.uber.org/zap"
)

var (
	crcTable = crc32.MakeTable(crc32.Castagnoli)
)

type WAL struct {
	lg *zap.Logger
	
	dir string // the living directory of the underlay files
	
	// dirFile is a fd for the wal directory for syncing on Rename
	dirFile *os.File
	
	metadata []byte           // metadata recorded at the head of each WAL
	state    raftpb.HardState // hardstate recorded at the head of WAL
	
	start     walpb.Snapshot // snapshot to start reading
	decoder   *decoder       // decoder to decode records
	readClose func() error   // closer for decode reader
	
	mu      sync.Mutex
	enti    uint64   // index of the last entry saved to the wal
	encoder *encoder // encoder to encode records
	
	locks []*fileutil.LockedFile // the locked files the WAL holds (the name is increasing)
	fp    *filePipeline
}

// Create creates a WAL ready for appending records. The given metadata is
// recorded at the head of each WAL file, and can be retrieved with ReadAll.
func Create(lg *zap.Logger, dirpath string, metadata []byte) (*WAL, error) {
	if Exist(dirpath) {
		return nil, os.ErrExist
	}
}
