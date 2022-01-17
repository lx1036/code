package raft

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	testPath      = "permTest"
	snapPath      = "snapshots"
	metaFilePath  = "meta.json"
	stateFilePath = "state.bin"
	tmpSuffix     = ".tmp"
)

// FileSnapshotStore implements the SnapshotStore interface and allows
// snapshots to be made on the local disk.
type FileSnapshotStore struct {
	path   string
	retain int

	// noSync, if true, skips crash-safe file fsync api calls.
	// It's a private field, only used in testing
	noSync bool
}

// NewFileSnapshotStore creates a new FileSnapshotStore based
// on a base directory. The `retain` parameter controls how many
// snapshots are retained. Must be at least 1.
func NewFileSnapshotStore(base string, retain int) (*FileSnapshotStore, error) {
	// Ensure our path exists
	path := filepath.Join(base, snapPath)
	if err := os.MkdirAll(path, 0777); err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("snapshot path:%s not accessible: %v", path, err)
	}

	return &FileSnapshotStore{
		path:   path,
		retain: retain,
	}, nil
}

func (store *FileSnapshotStore) Create(index, term uint64, configuration Configuration, configurationIndex uint64, trans Transport) (SnapshotSink, error) {
	panic("implement me")
}

func (store *FileSnapshotStore) List() ([]*SnapshotMeta, error) {
	snapshots, err := store.getSnapshots()
	if err != nil {
		klog.Errorf(fmt.Sprintf("failed to get snapshots error:%v", err))
		return nil, err
	}

	var snapMeta []*SnapshotMeta
	for _, meta := range snapshots {
		snapMeta = append(snapMeta, &meta.SnapshotMeta)
		if len(snapMeta) == store.retain {
			break
		}
	}
	return snapMeta, nil
}

// fileSnapshotMeta is stored on disk. We also put a CRC
// on disk so that we can verify the snapshot.
type fileSnapshotMeta struct {
	SnapshotMeta
	CRC []byte
}
type snapMetaSlice []*fileSnapshotMeta

func (s snapMetaSlice) Len() int {
	return len(s)
}
func (s snapMetaSlice) Less(i, j int) bool {
	if s[i].Term != s[j].Term {
		return s[i].Term < s[j].Term
	}
	if s[i].Index != s[j].Index {
		return s[i].Index < s[j].Index
	}
	return s[i].ID < s[j].ID
}
func (s snapMetaSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// getSnapshots returns all the known snapshots.
func (store *FileSnapshotStore) getSnapshots() ([]*fileSnapshotMeta, error) {
	// Get the eligible snapshots
	snapshots, err := ioutil.ReadDir(store.path)
	if err != nil {
		klog.Errorf(fmt.Sprintf("failed to scan snapshot directory err:%v", err))
		return nil, err
	}

	var snapMeta []*fileSnapshotMeta
	for _, snap := range snapshots {
		// Ignore any files
		if !snap.IsDir() {
			continue
		}

		// Ignore any temporary snapshots
		dirName := snap.Name()
		if strings.HasSuffix(dirName, tmpSuffix) {
			klog.Warningf(fmt.Sprintf("found temporary snapshot name:%s", dirName))
			continue
		}

		// Try to read the meta data
		meta, err := store.readMeta(dirName)
		if err != nil {
			klog.Warningf(fmt.Sprintf("failed to read metadata name:%s err:%v", dirName, err))
			continue
		}

		// Append, but only return up to the retain count
		snapMeta = append(snapMeta, meta)
	}

	// Sort the snapshot, reverse we get new -> old
	sort.Sort(sort.Reverse(snapMetaSlice(snapMeta)))

	return snapMeta, nil
}

// readMeta is used to read the meta data for a given named backup
func (store *FileSnapshotStore) readMeta(name string) (*fileSnapshotMeta, error) {
	// Open the meta file
	metaPath := filepath.Join(store.path, name, metaFilePath)
	file, err := os.Open(metaPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Buffer the file IO
	buffered := bufio.NewReader(file)

	// Read in the JSON
	meta := &fileSnapshotMeta{}
	dec := json.NewDecoder(buffered)
	if err := dec.Decode(meta); err != nil {
		return nil, err
	}
	return meta, nil
}

func (store *FileSnapshotStore) Open(id string) (*SnapshotMeta, io.ReadCloser, error) {
	panic("implement me")
}
