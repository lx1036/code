package wal

import (
	"errors"
	"fmt"
	"io"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"strings"

	"go.etcd.io/etcd/client/pkg/v3/fileutil"
	"go.etcd.io/etcd/server/v3/wal/walpb"
)

var errBadWALName = errors.New("bad wal name")

// Exist returns true if there are any files in a given directory.
func Exist(dir string) bool {
	names, err := fileutil.ReadDir(dir, fileutil.WithExt(".wal"))
	if err != nil {
		return false
	}
	return len(names) != 0
}

func openAtIndex(dirpath string, snap walpb.Snapshot, write bool) (*WAL, error) {
	names, nameIndex, err := selectWALFiles(dirpath, snap)
	if err != nil {
		return nil, err
	}

	rs, ls, closer, err := openWALFiles(dirpath, names, nameIndex, write)
	if err != nil {
		return nil, err
	}

	// create a WAL ready for reading
	w := &WAL{
		dir:       dirpath,
		start:     snap,
		decoder:   newDecoder(rs...),
		readClose: closer,
		locks:     ls,
	}

	if write {
		// write reuses the file descriptors from read; don't close so
		// WAL can append without dropping the file lock
		w.readClose = nil
		if _, _, err := parseWALName(filepath.Base(w.tail().Name())); err != nil {
			closer()
			return nil, err
		}
		w.fp = newFilePipeline(w.dir, SegmentSizeBytes)
	}

	return w, nil
}
func selectWALFiles(dirpath string, snap walpb.Snapshot) ([]string, int, error) {
	names, err := readWALNames(dirpath)
	if err != nil {
		return nil, -1, err
	}

	nameIndex, ok := searchIndex(names, snap.Index)
	if !ok || !isValidSeq(names[nameIndex:]) {
		err = ErrFileNotFound
		return nil, -1, err
	}

	return names, nameIndex, nil
}

// searchIndex returns the last array index of names whose raft index section is
// equal to or smaller than the given index.
// The given names MUST be sorted.
func searchIndex(names []string, index uint64) (int, bool) {
	for i := len(names) - 1; i >= 0; i-- {
		name := names[i]
		_, curIndex, err := parseWALName(name)
		if err != nil {
			klog.Fatalf(fmt.Sprintf("failed to parse WAL file name path:%s err:%v", name, err))
		}
		if index >= curIndex {
			return i, true
		}
	}
	return -1, false
}

// names should have been sorted based on sequence number.
// isValidSeq checks whether seq increases continuously.
func isValidSeq(names []string) bool {
	var lastSeq uint64
	for _, name := range names {
		curSeq, _, err := parseWALName(name)
		if err != nil {
			klog.Fatalf(fmt.Sprintf("failed to parse WAL file name path:%s err:%v", name, err))
		}
		if lastSeq != 0 && lastSeq != curSeq-1 {
			return false
		}
		lastSeq = curSeq
	}
	return true
}
func readWALNames(dirpath string) ([]string, error) {
	names, err := fileutil.ReadDir(dirpath)
	if err != nil {
		return nil, err
	}
	wnames := checkWalNames(names)
	if len(wnames) == 0 {
		return nil, ErrFileNotFound
	}
	return wnames, nil
}
func checkWalNames(names []string) []string {
	wnames := make([]string, 0)
	for _, name := range names {
		if _, _, err := parseWALName(name); err != nil {
			// don't complain about left over tmp files
			if !strings.HasSuffix(name, ".tmp") {
				klog.Errorf(fmt.Sprintf("ignored file in WAL directory path:%s err:%v", name, err))
			}
			continue
		}
		wnames = append(wnames, name)
	}
	return wnames
}
func openWALFiles(dirpath string, names []string, nameIndex int, write bool) ([]io.Reader, []*fileutil.LockedFile, func() error, error) {
	rcs := make([]io.ReadCloser, 0)
	rs := make([]io.Reader, 0)
	ls := make([]*fileutil.LockedFile, 0)
	for _, name := range names[nameIndex:] {
		p := filepath.Join(dirpath, name)
		if write {
			l, err := fileutil.TryLockFile(p, os.O_RDWR, fileutil.PrivateFileMode)
			if err != nil {
				closeAll(rcs...)
				return nil, nil, nil, err
			}
			ls = append(ls, l)
			rcs = append(rcs, l)
		} else {
			rf, err := os.OpenFile(p, os.O_RDONLY, fileutil.PrivateFileMode)
			if err != nil {
				closeAll(rcs...)
				return nil, nil, nil, err
			}
			ls = append(ls, nil)
			rcs = append(rcs, rf)
		}
		rs = append(rs, rcs[len(rcs)-1])
	}

	closer := func() error { return closeAll(rcs...) }

	return rs, ls, closer, nil
}

func closeAll(rcs ...io.ReadCloser) error {
	for _, f := range rcs {
		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
}

func parseWALName(str string) (seq, index uint64, err error) {
	if !strings.HasSuffix(str, ".wal") {
		return 0, 0, errBadWALName
	}
	_, err = fmt.Sscanf(str, "%016x-%016x.wal", &seq, &index)
	return seq, index, err
}

func walName(seq, index uint64) string {
	return fmt.Sprintf("%016x-%016x.wal", seq, index)
}
