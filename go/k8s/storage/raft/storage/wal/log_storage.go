package wal

import (
	"fmt"
	"k8s.io/klog"
	"os"
	"sort"
	"strings"
)

type logEntryStorage struct {
	storage *Storage

	dir string
}

// 打开 dir/xxx.log 文件
func (logEntry *logEntryStorage) open() error {
	// sort list dir/xxx.log 文件
	logFilenames, err := listLogEntryFiles(logEntry.dir)
	if err != nil {
		return err
	}

	// 没有log历史文件，则需要创建一个index=0的文件
	if len(logFilenames) == 0 {
		logEntry.createNew(1)
	}

	return nil
}

func (logEntry *logEntryStorage) createNew(index uint64) {

	f, err := createLogEntryFile()

}

func openLogStorage(dir string, storage *Storage) (*logEntryStorage, error) {
	logEntry := &logEntryStorage{
		storage: storage,
		dir:     dir,
	}

	if err := logEntry.open(); err != nil {
		return nil, err
	}

	return logEntry, nil
}
