package wal

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
)

type logEntryFile struct {
	file *os.File
	
}




func createLogEntryFile(dir string, name logFilename) (*logEntryFile, error) {
	filename := path.Join(dir, name.String()) // log entry 文件 dir/seq-index.log
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC|os.O_APPEND, 0600)
	if err != nil {
		return nil, err
	}
	
	l := &logEntryFile{
		file: file,
	}
	if err = l.OpenWrite(); err != nil {
		return nil, err
	}
	
	return l, nil
}




// 日志文件名字组成格式：seq-index.log
type logFilename struct {
	seq   uint64 // 文件序号
	index uint64 // log entry index
}

func (logFile *logFilename) Parse(filename string) bool {
	_, err := fmt.Sscanf(filename, "%016x-%016x.log", &logFile.seq, &logFile.index)
	return err == nil
}

func (logFile *logFilename) String() string {
	return fmt.Sprintf("%016x-%016x.log", logFile.seq, logFile.index)
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

// sort list dir 目录下 xxx.log 文件
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
	klog.Infof(fmt.Sprintf("[listLogEntryFiles]%s", strings.Join(names, ",")))
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
