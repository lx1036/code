package wal

import (
	"bytes"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/etcd/raft"

	"go.etcd.io/etcd/client/pkg/v3/fileutil"
	"go.etcd.io/etcd/pkg/v3/pbutil"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/server/v3/wal/walpb"
	"k8s.io/klog/v2"
)

const (
	// 所有的日志类型如下所示，会被append追加存储在wal文件中
	metadataType int64 = iota + 1
	entryType
	stateType
	crcType
	snapshotType

	// warnSyncDuration is the amount of time allotted to an fsync before
	// logging a warning
	warnSyncDuration = time.Second
)

var (
	crcTable = crc32.MakeTable(crc32.Castagnoli)

	// SegmentSizeBytes is the preallocated size of each wal segment file.
	// The actual size might be larger than this. In general, the default
	// value should be used, but this is defined as an exported variable
	// so that tests can set a different segment size.
	SegmentSizeBytes int64 = 64 * 1000 * 1000 // 64MB

	ErrFileNotFound     = errors.New("wal: file not found")
	ErrSnapshotNotFound = errors.New("wal: snapshot not found")
	ErrMetadataConflict = errors.New("wal: conflicting metadata found")
	ErrCRCMismatch      = errors.New("wal: crc mismatch")
	ErrSnapshotMismatch = errors.New("wal: snapshot mismatch")
)

// WAL用于以append记录的方式快速记录数据到持久化存储中
// 只可能处于append模式或者读模式，但是不能同时处于两种模式中
// 新创建的WAL处于append模式，刚打开的WAL处于读模式
type WAL struct {
	// 存放WAL文件的目录
	dir string // the living directory of the underlay files

	// 根据前面的dir成员创建的File指针
	// dirFile is a fd for the wal directory for syncing on Rename
	dirFile *os.File

	// 每个WAL文件最开始的地方，都需要写入的元数据
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

func (w *WAL) Save(hardState raftpb.HardState, entries []raftpb.Entry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// short cut, do not call sync
	if raft.IsEmptyHardState(hardState) && len(entries) == 0 {
		return nil
	}

	mustSync := raft.MustSync(hardState, w.state, len(entries))

	for i := range entries {
		if err := w.saveEntry(&entries[i]); err != nil {
			return err
		}
	}
	if err := w.saveState(&hardState); err != nil {
		return err
	}

	curOff, err := w.tail().Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	if curOff < SegmentSizeBytes {
		if mustSync {
			return w.sync()
		}
		return nil
	}

	return w.cut()
}

func (w *WAL) saveEntry(e *raftpb.Entry) error {
	b := pbutil.MustMarshal(e)
	rec := &walpb.Record{Type: entryType, Data: b}
	if err := w.encoder.encode(rec); err != nil {
		return err
	}
	w.enti = e.Index
	return nil
}

func (w *WAL) saveState(s *raftpb.HardState) error {
	if raft.IsEmptyHardState(*s) {
		return nil
	}
	w.state = *s
	b := pbutil.MustMarshal(s)
	rec := &walpb.Record{Type: stateType, Data: b}
	return w.encoder.encode(rec)
}

// cut closes current file written and creates a new one ready to append.
// cut first creates a temp wal file and writes necessary headers into it.
// Then cut atomically rename temp wal file to a wal file.
func (w *WAL) cut() error {
	// close old wal file; truncate to avoid wasting space if an early cut
	off, serr := w.tail().Seek(0, io.SeekCurrent)
	if serr != nil {
		return serr
	}

	if err := w.tail().Truncate(off); err != nil {
		return err
	}

	if err := w.sync(); err != nil {
		return err
	}

	fpath := filepath.Join(w.dir, walName(w.seq()+1, w.enti+1))

	// create a temp wal file with name sequence + 1, or truncate the existing one
	newTail, err := w.fp.Open()
	if err != nil {
		return err
	}

	// update writer and save the previous crc
	w.locks = append(w.locks, newTail)
	prevCrc := w.encoder.crc.Sum32()
	w.encoder, err = newFileEncoder(w.tail().File, prevCrc)
	if err != nil {
		return err
	}

	if err = w.saveCrc(prevCrc); err != nil {
		return err
	}

	if err = w.encoder.encode(&walpb.Record{Type: metadataType, Data: w.metadata}); err != nil {
		return err
	}

	if err = w.saveState(&w.state); err != nil {
		return err
	}

	// atomically move temp wal file to wal file
	if err = w.sync(); err != nil {
		return err
	}

	off, err = w.tail().Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	if err = os.Rename(newTail.Name(), fpath); err != nil {
		return err
	}
	start := time.Now()
	if err = fileutil.Fsync(w.dirFile); err != nil {
		return err
	}
	walFsyncSec.Observe(time.Since(start).Seconds())

	// reopen newTail with its new path so calls to Name() match the wal filename format
	newTail.Close()

	if newTail, err = fileutil.LockFile(fpath, os.O_WRONLY, fileutil.PrivateFileMode); err != nil {
		return err
	}
	if _, err = newTail.Seek(off, io.SeekStart); err != nil {
		return err
	}

	w.locks[len(w.locks)-1] = newTail

	prevCrc = w.encoder.crc.Sum32()
	w.encoder, err = newFileEncoder(w.tail().File, prevCrc)
	if err != nil {
		return err
	}

	klog.Infof(fmt.Sprintf("created a new WAL segment path:%s", fpath))
	return nil
}

func (w *WAL) saveCrc(prevCrc uint32) error {
	return w.encoder.encode(&walpb.Record{Type: crcType, Crc: prevCrc})
}

func (w *WAL) SaveSnapshot(e walpb.Snapshot) error {
	if err := walpb.ValidateSnapshotForWrite(&e); err != nil {
		return err
	}

	b := pbutil.MustMarshal(&e)

	w.mu.Lock()
	defer w.mu.Unlock()

	rec := &walpb.Record{Type: snapshotType, Data: b}
	if err := w.encoder.encode(rec); err != nil {
		return err
	}
	// update enti only when snapshot is ahead of last index
	if w.enti < e.Index {
		w.enti = e.Index
	}

	return w.sync()
}

// 系统调用把snapshot写入持久化磁盘里
func (w *WAL) sync() error {
	if w.encoder != nil {
		if err := w.encoder.flush(); err != nil {
			return err
		}
	}
	start := time.Now()
	err := fileutil.Fdatasync(w.tail().File)

	took := time.Since(start)
	if took > warnSyncDuration {
		klog.Errorf(fmt.Sprintf("sync duration of %v, expected less than %v", took, warnSyncDuration))
	}

	walFsyncSec.Observe(took.Seconds())

	return err
}

func (w *WAL) tail() *fileutil.LockedFile {
	if len(w.locks) > 0 {
		return w.locks[len(w.locks)-1]
	}
	return nil
}

// walName: %016x-%016x.wal, seq-index
func (w *WAL) seq() uint64 {
	t := w.tail()
	if t == nil {
		return 0
	}
	seq, _, err := parseWALName(filepath.Base(t.Name()))
	if err != nil {
		klog.Fatalf(fmt.Sprintf("failed to parse WAL name:%s err:%v", t.Name(), err))
	}
	return seq
}

func (w *WAL) renameWAL(tmpdirpath string) (*WAL, error) {
	if err := os.RemoveAll(w.dir); err != nil {
		return nil, err
	}
	// On non-Windows platforms, hold the lock while renaming. Releasing
	// the lock and trying to reacquire it quickly can be flaky because
	// it's possible the process will fork to spawn a process while this is
	// happening. The fds are set up as close-on-exec by the Go runtime,
	// but there is a window between the fork and the exec where another
	// process holds the lock.
	if err := os.Rename(tmpdirpath, w.dir); err != nil {
		if _, ok := err.(*os.LinkError); ok {
			return w.renameWALUnlock(tmpdirpath)
		}
		return nil, err
	}
	w.fp = newFilePipeline(w.dir, SegmentSizeBytes)
	df, err := fileutil.OpenDir(w.dir)
	w.dirFile = df
	return w, err
}

func (w *WAL) renameWALUnlock(tmpdirpath string) (*WAL, error) {
	// rename of directory with locked files doesn't work on windows/cifs;
	// close the WAL to release the locks so the directory can be renamed.
	klog.Infof(fmt.Sprintf("closing WAL to release flock and retry directory renaming from %s to %s", tmpdirpath, w.dir))
	w.Close()

	if err := os.Rename(tmpdirpath, w.dir); err != nil {
		return nil, err
	}

	// reopen and relock
	newWAL, oerr := Open(w.dir, walpb.Snapshot{})
	if oerr != nil {
		return nil, oerr
	}
	if _, _, _, err := newWAL.ReadAll(); err != nil {
		newWAL.Close()
		return nil, err
	}
	return newWAL, nil
}

// ReadAll ReadAll函数负责从当前WAL实例中读取所有的记录
// 如果是可写模式，那么必须独处所有的记录，否则将报错
// 如果是只读模式，将尝试读取所有的记录，但是如果读出的记录没有满足快照数据的要求，将返回ErrSnapshotNotFound
// 而如果读出来的快照数据与要求的快照数据不匹配，返回所有的记录以及ErrSnapshotMismatch
// ReadAll reads out records of the current WAL.
// If opened in write mode, it must read out all records until EOF. Or an error
// will be returned.
// If opened in read mode, it will try to read all records if possible.
// If it cannot read out the expected snap, it will return ErrSnapshotNotFound.
// If loaded snap doesn't match with the expected one, it will return
// all the records and error ErrSnapshotMismatch.
// INFO: detect not-last-snap error.
// INFO: maybe loose the checking of match.
// After ReadAll, the WAL will be ready for appending new records.
func (w *WAL) ReadAll() (metadata []byte, state raftpb.HardState, ents []raftpb.Entry, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	rec := &walpb.Record{}
	decoder := w.decoder

	var match bool
	// 循环读出record记录
	for err = decoder.decode(rec); err == nil; err = decoder.decode(rec) {
		switch rec.Type {
		case entryType:
			e := mustUnmarshalEntry(rec.Data)
			if e.Index > w.start.Index {
				ents = append(ents[:e.Index-w.start.Index-1], e)
			}
			w.enti = e.Index

		case stateType:
			state = mustUnmarshalState(rec.Data)

		case metadataType:
			if metadata != nil && !bytes.Equal(metadata, rec.Data) {
				state.Reset()
				return nil, state, nil, ErrMetadataConflict
			}
			metadata = rec.Data

		case crcType:
			crc := decoder.crc.Sum32()
			// current crc of decoder must match the crc of the record.
			// do no need to match 0 crc, since the decoder is a new one at this case.
			if crc != 0 && rec.Validate(crc) != nil {
				state.Reset()
				return nil, state, nil, ErrCRCMismatch
			}
			decoder.updateCRC(rec.Crc)

		case snapshotType:
			var snap walpb.Snapshot
			pbutil.MustUnmarshal(&snap, rec.Data)
			if snap.Index == w.start.Index {
				if snap.Term != w.start.Term {
					state.Reset()
					return nil, state, nil, ErrSnapshotMismatch
				}
				match = true
			}

		default:
			state.Reset()
			return nil, state, nil, fmt.Errorf("unexpected block type %d", rec.Type)
		}
	}

	// 到了这里，读取WAL日志的循环就结束了，中间可能出错
	// 如果读完了所有文件，则err = io.EOF
	// 如果是没有读完文件而中间有其他错误，那么err就不是EOF了，下面会分别处理只读模式和写模式
	switch w.tail() {
	case nil:
		// We do not have to read out all entries in read mode.
		// The last record maybe a partial written one, so
		// ErrunexpectedEOF might be returned.
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			state.Reset()
			return nil, state, nil, err
		}
	default:
		// We must read all of the entries if WAL is opened in write mode.
		if err != io.EOF {
			state.Reset()
			return nil, state, nil, err
		}
		// decodeRecord() will return io.EOF if it detects a zero record,
		// but this zero record may be followed by non-zero records from
		// a torn write. Overwriting some of these non-zero records, but
		// not all, will cause CRC errors on WAL open. Since the records
		// were never fully synced to disk in the first place, it's safe
		// to zero them out to avoid any CRC errors from new writes.
		if _, err = w.tail().Seek(w.decoder.lastOffset(), io.SeekStart); err != nil {
			return nil, state, nil, err
		}
		if err = fileutil.ZeroToEnd(w.tail().File); err != nil {
			return nil, state, nil, err
		}
	}

	err = nil
	if !match {
		err = ErrSnapshotNotFound
	}

	// close decoder, disable reading
	if w.readClose != nil {
		w.readClose()
		w.readClose = nil
	}
	w.start = walpb.Snapshot{}

	w.metadata = metadata

	if w.tail() != nil {
		// create encoder (chain crc with the decoder), enable appending
		w.encoder, err = newFileEncoder(w.tail().File, w.decoder.lastCRC())
		if err != nil {
			return
		}
	}
	w.decoder = nil

	return metadata, state, ents, err
}

// Close closes the current WAL file and directory.
func (w *WAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.fp != nil {
		w.fp.Close()
		w.fp = nil
	}

	if w.tail() != nil {
		if err := w.sync(); err != nil {
			return err
		}
	}
	for _, l := range w.locks {
		if l == nil {
			continue
		}
		if err := l.Close(); err != nil {
			klog.Errorf(fmt.Sprintf("failed to unlock during closing wal: %v", err))
		}
	}

	return w.dirFile.Close()
}

func (w *WAL) cleanupWAL() {
	var err error
	if err = w.Close(); err != nil {
		klog.Fatalf(fmt.Sprintf("failed to close WAL during cleanup err:%v", err))
	}
	brokenDirName := fmt.Sprintf("%s.broken.%v", w.dir, time.Now().Format("20060102.150405.999999"))
	if err = os.Rename(w.dir, brokenDirName); err != nil {
		klog.Fatalf(fmt.Sprintf("failed to rename WAL during cleanup source-path:%s rename-path:%s err:%v", w.dir, brokenDirName, err))
	}
}

// Create creates a WAL ready for appending records. The given metadata is
// recorded at the head of each WAL file, and can be retrieved with ReadAll.
func Create(dataDir string, metadata []byte) (*WAL, error) {
	if Exist(dataDir) {
		return nil, os.ErrExist
	}
	// keep temporary wal directory so WAL initialization appears atomic
	tmpdirpath := filepath.Clean(dataDir) + ".tmp"
	if fileutil.Exist(tmpdirpath) {
		if err := os.RemoveAll(tmpdirpath); err != nil {
			return nil, err
		}
	}
	defer os.RemoveAll(tmpdirpath)

	if err := fileutil.CreateDirAll(tmpdirpath); err != nil {
		klog.Errorf(fmt.Sprintf("[WAL Create]failed to create a temporary WAL directory tmp-dir-path:%s dir-path:%s err:%v", tmpdirpath, dataDir, err))
		return nil, err
	}
	path := filepath.Join(tmpdirpath, walName(0, 0))
	f, err := fileutil.LockFile(path, os.O_WRONLY|os.O_CREATE, fileutil.PrivateFileMode)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[WAL Create]failed to flock an initial WAL file path:%s err:%v", path, err))
		return nil, err
	}
	if _, err = f.Seek(0, io.SeekEnd); err != nil {
		klog.Errorf(fmt.Sprintf("[WAL Create]failed to seek an initial WAL file path:%s err:%v", path, err))
		return nil, err
	}
	if err = fileutil.Preallocate(f.File, SegmentSizeBytes, true); err != nil {
		klog.Errorf(fmt.Sprintf("[WAL Create]failed to preallocate an initial WAL file path:%s segment-bytes:%d err:%v", path, SegmentSizeBytes, err))
		return nil, err
	}

	// INFO: crcType -> metadataType -> snapshotType -> entryType -> stateType
	w := &WAL{
		dir:      dataDir,
		metadata: metadata,
	}
	w.encoder, err = newFileEncoder(f.File, 0)
	if err != nil {
		return nil, err
	}
	w.locks = append(w.locks, f)
	if err = w.saveCrc(0); err != nil {
		return nil, err
	}
	if err = w.encoder.encode(&walpb.Record{Type: metadataType, Data: metadata}); err != nil {
		return nil, err
	}
	if err = w.SaveSnapshot(walpb.Snapshot{}); err != nil {
		return nil, err
	}

	if w, err = w.renameWAL(tmpdirpath); err != nil {
		klog.Errorf(fmt.Sprintf("[WAL Create]failed to rename the temporary WAL directory tmp-dir-path:%s dir-path:%s err:%v", tmpdirpath, w.dir, err))
		return nil, err
	}

	defer func() {
		if err != nil {
			w.cleanupWAL()
		}
	}()

	// directory was renamed; sync parent dir to persist rename
	parentDir, err := fileutil.OpenDir(filepath.Dir(w.dir))
	if err != nil {
		klog.Errorf(fmt.Sprintf("[WAL Create]failed to open the parent data directory parent-dir-path:%s dir-path:%s err:%v", filepath.Dir(w.dir), w.dir, err))
		return nil, err
	}
	defer parentDir.Close()
	if err = fileutil.Fsync(parentDir); err != nil {
		klog.Errorf(fmt.Sprintf("[WAL Create]failed to fsync the parent data directory file parent-dir-path:%s dir-path:%s err:%v", filepath.Dir(w.dir), w.dir, err))
		return nil, err
	}

	return w, nil
}

// Open opens the WAL at the given snap.
// The snap SHOULD have been previously saved to the WAL, or the following
// ReadAll will fail.
// The returned WAL is ready to read and the first record will be the one after
// the given snap. The WAL cannot be appended to before reading out all of its
// previous records.
func Open(dirpath string, snap walpb.Snapshot) (*WAL, error) {
	w, err := openAtIndex(dirpath, snap, true)
	if err != nil {
		return nil, err
	}
	if w.dirFile, err = fileutil.OpenDir(w.dir); err != nil {
		return nil, err
	}
	return w, nil
}
