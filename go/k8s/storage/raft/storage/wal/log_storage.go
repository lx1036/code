package wal

import (
	"fmt"
	"os"
	"sort"
)

type logEntryStorage struct {
	storage *Storage
	
	dir string
}

// 打开 dir/xxx.log 文件
func (logEntry *logEntryStorage) open() error {
	// sort list dir/xxx.log 文件
	
	return nil
}

func openLogStorage(dir string, storage *Storage) (*logEntryStorage, error) {
	logEntry := &logEntryStorage{
		storage: storage,
		dir: dir,
	}
	
	if err := logEntry.open(); err != nil {
		return nil, err
	}
	
	return logEntry, nil
}


// 日志文件名字组成格式：seq-index.log
type logFilename struct {
	seq uint64 // 文件序号
	index uint64 // log entry index
}

func (logFile *logFilename) Parse(filename string) bool {
	_, err := fmt.Sscanf(filename, "%016x-%016x.log", &logFile.seq, &logFile.index)
	return err == nil
}

type logFilenameSlice []logFilename

func (l logFilenameSlice) Len() int {
	return len(l)
}

func (l logFilenameSlice) Less(i, j int) bool {
	return l[i].seq < l[j].seq
}

func (l logFilenameSlice) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func listLogEntryFiles(dir string) ([]logFilename, error) {
	file, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	names, err := file.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	var logFilenames []logFilename
	for _, name := range names {
		var filename logFilename
		if filename.Parse(name) {
			logFilenames = append(logFilenames, filename)
		}
	}
	
	sort.Sort(logFilenameSlice(logFilenames))
	
	return logFilenames, nil
}
