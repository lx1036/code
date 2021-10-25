package wal

import (
	"k8s-lx1036/k8s/storage/raft/proto"
)

type Storage struct {
	config *Config

	// 记录压缩
	truncateIndex uint64
	truncateTerm  uint64

	hardState proto.HardState

	// META文件对象
	metaFile *metaFile

	prevCommit uint64 // 有commit变化时sync一下

	// 最重要的字段，log entry
	logEntry *logEntryStorage
}

func (s Storage) InitialState() (proto.HardState, error) {
	panic("implement me")
}

func (s Storage) Entries(lo, hi uint64, maxSize uint64) (entries []*proto.Entry, isCompact bool, err error) {
	panic("implement me")
}

func (s Storage) Term(i uint64) (term uint64, isCompact bool, err error) {
	panic("implement me")
}

func (s Storage) FirstIndex() (uint64, error) {
	panic("implement me")
}

func (s Storage) LastIndex() (uint64, error) {
	panic("implement me")
}

func (s Storage) StoreEntries(entries []*proto.Entry) error {
	panic("implement me")
}

func (s Storage) StoreHardState(st proto.HardState) error {
	panic("implement me")
}

func (s Storage) Truncate(index uint64) error {
	panic("implement me")
}

func (s Storage) ApplySnapshot(meta proto.SnapshotMeta) error {
	panic("implement me")
}

func (s Storage) Close() {
	panic("implement me")
}

func NewStorage(dir string, config *Config) (*Storage, error) {
	// TODO: check dir is a dir mode

	mf, hardState, meta, err := openMetaFile(dir) // 打开 dir/META 文件，存储的是压缩相关的元信息
	if err != nil {
		return nil, err
	}

	storage := &Storage{
		config:        config,
		truncateIndex: meta.truncateIndex,
		truncateTerm:  meta.truncateTerm,
		hardState:     hardState,
		metaFile:      mf,
		prevCommit:    hardState.Commit,
	}

	// 加载日志文件
	logEntry, err := openLogStorage(dir, storage)
	if err != nil {
		return nil, err
	}
	storage.logEntry = logEntry

	return storage, nil
}
