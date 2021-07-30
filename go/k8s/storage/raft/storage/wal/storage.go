package wal

import (
	"github.com/tiglabs/raft/proto"
)

type Storage struct {
	config *Config
	
	// 记录压缩
	truncateIndex uint64
	truncateTerm uint64
	
	hardState proto.HardState
	
	// META文件对象
	metaFile *metaFile
	
	prevCommit uint64 // 有commit变化时sync一下
	
	// 最重要的字段，log entry
	logEntry *logEntryStorage
}





func NewStorage(dir string, config *Config) (*Storage, error) {
	// TODO: check dir is a dir mode
	
	mf, hardState, meta, err := openMetaFile(dir) // 打开 dir/META 文件，存储的是压缩相关的元信息
	if err != nil {
		return nil, err
	}
	
	storage := &Storage{
		config: config,
		truncateIndex: meta.truncateIndex,
		truncateTerm: meta.truncateTerm,
		hardState: hardState,
		metaFile: mf,
		prevCommit: hardState.Commit,
	}
	
	// 加载日志文件
	logEntry, err := openLogStorage(dir, storage)
	if err != nil {
		return nil, err
	}
	storage.logEntry = logEntry
	
	
	return storage, nil
}
