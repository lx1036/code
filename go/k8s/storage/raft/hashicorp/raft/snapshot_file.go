package raft

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"hash"
	"hash/crc64"
	"io"
	"io/ioutil"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	testPath      = "permTest"
	snapPath      = "snapshots"
	metaFilePath  = "meta.json"
	stateFilePath = "fsm_snapshot.bin"
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

// bufferedFile is returned when we open a snapshot. This way
// reads are buffered and the file still gets closed.
type bufferedFile struct {
	bh *bufio.Reader
	fh *os.File
}

func (b *bufferedFile) Read(p []byte) (n int, err error) {
	return b.bh.Read(p)
}

func (b *bufferedFile) Close() error {
	return b.fh.Close()
}

// Open INFO: restore from snapshot file and apply into fsm, @see Create()
func (store *FileSnapshotStore) Open(id string) (*SnapshotMeta, io.ReadCloser, error) {
	// Get the metadata
	meta, err := store.readMeta(id)
	if err != nil {
		klog.Errorf(fmt.Sprintf("failed to get meta data to open snapshot err:%v", err))
		return nil, nil, err
	}

	// Open the state file
	statePath := filepath.Join(store.path, id, stateFilePath)
	fh, err := os.Open(statePath)
	if err != nil {
		klog.Errorf(fmt.Sprintf("failed to open state file err:%v", err))
		return nil, nil, err
	}

	// Verify the hash
	stateHash := crc64.New(crc64.MakeTable(crc64.ECMA))
	_, err = io.Copy(stateHash, fh)
	if err != nil {
		klog.Errorf(fmt.Sprintf("failed to read state file err:%v", err))
		fh.Close()
		return nil, nil, err
	}
	computed := stateHash.Sum(nil)
	if bytes.Compare(meta.CRC, computed) != 0 {
		klog.Errorf(fmt.Sprintf("CRC checksum failed stored:%s computed:%s", meta.CRC, computed))
		fh.Close()
		return nil, nil, fmt.Errorf("CRC mismatch")
	}

	// Seek to the start
	if _, err := fh.Seek(0, 0); err != nil {
		klog.Errorf(fmt.Sprintf("state file seek failed err:%v", err))
		fh.Close()
		return nil, nil, err
	}
	// Return a buffered file
	buffered := &bufferedFile{
		bh: bufio.NewReader(fh),
		fh: fh,
	}

	return &meta.SnapshotMeta, buffered, nil
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

// FileSnapshotSink implements SnapshotSink with a file.
type FileSnapshotSink struct {
	store     *FileSnapshotStore
	dir       string
	parentDir string
	meta      fileSnapshotMeta

	noSync bool

	stateFile *os.File
	stateHash hash.Hash64
	buffered  *bufio.Writer

	closed bool
}

// Create INFO: snapshot fsm and persist data into sink, @see Open()
func (store *FileSnapshotStore) Create(index, term uint64, configuration Configuration, configurationIndex uint64) (SnapshotSink, error) {
	// Create a new path
	name := snapshotName(term, index)
	path := filepath.Join(store.path, name+tmpSuffix)
	klog.Infof(fmt.Sprintf("creating new snapshot path:%s", path))

	// Make the directory
	if err := os.MkdirAll(path, 0755); err != nil {
		klog.Errorf(fmt.Sprintf("failed to make snapshot directly err:%v", err))
		return nil, err
	}

	// Create the sink
	sink := &FileSnapshotSink{
		store:     store,
		dir:       path,
		parentDir: store.path,
		noSync:    store.noSync,
		meta: fileSnapshotMeta{
			SnapshotMeta: SnapshotMeta{
				ID:                 name,
				Index:              index,
				Term:               term,
				Configuration:      configuration,
				ConfigurationIndex: configurationIndex,
			},
			CRC: nil,
		},
	}

	// Write out the meta data
	if err := sink.writeMeta(); err != nil {
		klog.Errorf(fmt.Sprintf("failed to write metadata err:%v", err))
		return nil, err
	}

	// Open the state file
	statePath := filepath.Join(path, stateFilePath)
	fh, err := os.Create(statePath)
	if err != nil {
		klog.Errorf(fmt.Sprintf("failed to create state file err:%v", err))
		return nil, err
	}
	sink.stateFile = fh

	// Create a CRC64 hash
	sink.stateHash = crc64.New(crc64.MakeTable(crc64.ECMA))

	// INFO: Wrap both the hash and file in a MultiWriter with buffering
	//  该方法可以借鉴：同时写 file and hash, hash 会记录在 meta json file
	multi := io.MultiWriter(sink.stateFile, sink.stateHash)
	sink.buffered = bufio.NewWriter(multi)

	// Done
	return sink, nil
}

// writeMeta is used to write out the metadata we have.
func (s *FileSnapshotSink) writeMeta() error {
	var err error
	// Open the meta file
	metaPath := filepath.Join(s.dir, metaFilePath)
	var fh *os.File
	fh, err = os.Create(metaPath)
	if err != nil {
		return err
	}
	defer fh.Close()

	// Buffer the file IO
	buffered := bufio.NewWriter(fh)

	// Write out as JSON
	enc := json.NewEncoder(buffered)
	if err = enc.Encode(&s.meta); err != nil {
		return err
	}

	if err = buffered.Flush(); err != nil {
		return err
	}

	if !s.noSync {
		if err = fh.Sync(); err != nil {
			return err
		}
	}

	return nil
}

// Write is used to append to the state file. We write to the
// buffered IO object to reduce the amount of context switches.
func (s *FileSnapshotSink) Write(b []byte) (int, error) {
	return s.buffered.Write(b)
}

// ID returns the ID of the snapshot, can be used with Open()
// after the snapshot is finalized.
func (s *FileSnapshotSink) ID() string {
	return s.meta.ID
}

// Close persist snapshot file and meta file
func (s *FileSnapshotSink) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true

	// Close the open handles
	if err := s.finalize(); err != nil {
		klog.Errorf(fmt.Sprintf("failed to finalize snapshot err:%v", err))
		if delErr := os.RemoveAll(s.dir); delErr != nil {
			klog.Errorf(fmt.Sprintf("failed to delete temporary snapshot directory path:%s err:%v", s.dir, delErr))
			return delErr
		}
		return err
	}

	// Write out the meta data
	if err := s.writeMeta(); err != nil {
		klog.Errorf(fmt.Sprintf("failed to write metadata err:%v", err))
		return err
	}

	// Move the directory into place
	newPath := strings.TrimSuffix(s.dir, tmpSuffix)
	if err := os.Rename(s.dir, newPath); err != nil {
		klog.Errorf(fmt.Sprintf("failed to move snapshot into place err:%v", err))
		return err
	}

}

func (s *FileSnapshotSink) Cancel() error {
	panic("implement me")
}

// snapshotName generates a name for the snapshot.
func snapshotName(term, index uint64) string {
	now := time.Now()
	msec := now.UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%d-%d-%d", term, index, msec)
}
