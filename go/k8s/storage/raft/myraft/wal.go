package myraft

import (
	"io/ioutil"
	"os"
	"sync"
)

type Option struct {
	SegmentNum  int
	SegmentSize int
	IsSync      bool // flush sync after append, 追加写后立刻从内存持久化到磁盘
}

var defaultOption = Option{
	SegmentNum:  2,
	SegmentSize: 20 * 1024 * 1024, // 20 MB
	IsSync:      false,
}

// INFO: 支持读写并发
type Wal struct {
	sync.RWMutex
	option *Option

	dir string
}

func NewWal(dir string, option *Option) (*Wal, error) {
	if option == nil {
		option = &defaultOption
	}

	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return nil, err
	}
	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, info := range fileInfos {
		if info.IsDir() {
			continue
		}
		name := info.Name()
		RecoverSegment(name, dir)
	}

	w := &Wal{
		option: option,
		dir:    dir,
	}

	return w, nil
}
