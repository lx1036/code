//go:build linux
// +build linux

package bpf

import (
	"fmt"
	"github.com/cilium/cilium/pkg/datapath/linux/probes"
	"golang.org/x/sys/unix"
	"io"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"unsafe"
)

const (
	// BPF syscall command constants. Must match enum bpf_cmd from linux/bpf.h
	BPF_MAP_CREATE          = 0
	BPF_MAP_LOOKUP_ELEM     = 1
	BPF_MAP_UPDATE_ELEM     = 2
	BPF_MAP_DELETE_ELEM     = 3
	BPF_MAP_GET_NEXT_KEY    = 4
	BPF_PROG_LOAD           = 5
	BPF_OBJ_PIN             = 6
	BPF_OBJ_GET             = 7
	BPF_PROG_ATTACH         = 8
	BPF_PROG_DETACH         = 9
	BPF_PROG_TEST_RUN       = 10
	BPF_PROG_GET_NEXT_ID    = 11
	BPF_MAP_GET_NEXT_ID     = 12
	BPF_PROG_GET_FD_BY_ID   = 13
	BPF_MAP_GET_FD_BY_ID    = 14
	BPF_OBJ_GET_INFO_BY_FD  = 15
	BPF_PROG_QUERY          = 16
	BPF_RAW_TRACEPOINT_OPEN = 17
	BPF_BTF_LOAD            = 18
	BPF_BTF_GET_FD_BY_ID    = 19
	BPF_TASK_FD_QUERY       = 20

	// BPF syscall attach types
	BPF_CGROUP_INET_INGRESS      = 0
	BPF_CGROUP_INET_EGRESS       = 1
	BPF_CGROUP_INET_SOCK_CREATE  = 2
	BPF_CGROUP_SOCK_OPS          = 3
	BPF_SK_SKB_STREAM_PARSER     = 4
	BPF_SK_SKB_STREAM_VERDICT    = 5
	BPF_CGROUP_DEVICE            = 6
	BPF_SK_MSG_VERDICT           = 7
	BPF_CGROUP_INET4_BIND        = 8
	BPF_CGROUP_INET6_BIND        = 9
	BPF_CGROUP_INET4_CONNECT     = 10
	BPF_CGROUP_INET6_CONNECT     = 11
	BPF_CGROUP_INET4_POST_BIND   = 12
	BPF_CGROUP_INET6_POST_BIND   = 13
	BPF_CGROUP_UDP4_SENDMSG      = 14
	BPF_CGROUP_UDP6_SENDMSG      = 15
	BPF_LIRC_MODE2               = 16
	BPF_FLOW_DISSECTOR           = 17
	BPF_CGROUP_SYSCTL            = 18
	BPF_CGROUP_UDP4_RECVMSG      = 19
	BPF_CGROUP_UDP6_RECVMSG      = 20
	BPF_CGROUP_INET4_GETPEERNAME = 29
	BPF_CGROUP_INET6_GETPEERNAME = 30
	BPF_CGROUP_INET4_GETSOCKNAME = 31
	BPF_CGROUP_INET6_GETSOCKNAME = 32

	// Flags for BPF_MAP_UPDATE_ELEM. Must match values from linux/bpf.h
	BPF_ANY     = 0
	BPF_NOEXIST = 1
	BPF_EXIST   = 2

	// Flags for BPF_MAP_CREATE. Must match values from linux/bpf.h
	BPF_F_NO_PREALLOC   = 1 << 0
	BPF_F_NO_COMMON_LRU = 1 << 1
	BPF_F_NUMA_NODE     = 1 << 2

	// Flags for BPF_PROG_QUERY
	BPF_F_QUERY_EFFECTVE = 1 << 0

	// Flags for accessing BPF object
	BPF_F_RDONLY = 1 << 3
	BPF_F_WRONLY = 1 << 4

	// Flag for stack_map, store build_id+offset instead of pointer
	BPF_F_STACK_BUILD_ID = 1 << 5
)

var (
	preAllocateMapSetting uint32 = BPF_F_NO_PREALLOC
)

// OpenOrCreateMap INFO: open or create a map file like cgroup in /sys/fs/bpf
func OpenOrCreateMap(path string, mapType MapType, keySize, valueSize, maxEntries,
	flags uint32, innerID uint32, pin bool) (int, bool, error) {
	var fd int

	redo := false
	isNewMap := false

recreate: // 新建一个 bpf map
	if _, err := os.Stat(path); os.IsNotExist(err) || redo {
		// mkdir dir
		mapDir := filepath.Dir(path)
		if _, err = os.Stat(mapDir); os.IsNotExist(err) {
			if err = os.MkdirAll(mapDir, 0755); err != nil {
				return 0, isNewMap, &os.PathError{
					Op:   "Unable create map base directory",
					Path: path,
					Err:  err,
				}
			}
		}

		fd, err = CreateMap(
			mapType,
			keySize,
			valueSize,
			maxEntries,
			flags,
			innerID,
			path,
		)

		defer func() {
			if err != nil {
				// In case of error, we need to close
				// this fd since it was open by CreateMap
				ObjClose(fd)
			}
		}()

		isNewMap = true

		if err != nil {
			return 0, isNewMap, err
		}

		if pin {
			err = ObjPin(fd, path)
			if err != nil {
				return 0, isNewMap, err
			}
		}

		return fd, isNewMap, nil
	}

	// 已经存在，但是如果需要重建
	fd, err := ObjGet(path)
	if err == nil {
		redo = objCheck(
			fd,
			path,
			mapType,
			keySize,
			valueSize,
			maxEntries,
			flags,
		)
		if redo == true {
			ObjClose(fd)
			goto recreate
		}
	}

	return fd, isNewMap, err
}

// CreateMap creates a Map of type mapType, with key size keySize, a value size of
// valueSize and the maximum amount of entries of maxEntries.
// mapType should be one of the bpf_map_type in linux kernel "uapi/linux/bpf.h"
// When mapType is the type HASH_OF_MAPS an innerID is required to point at a
// map fd which has the same type/keySize/valueSize/maxEntries as expected map
// entries. For all other mapTypes innerID is ignored and should be zeroed.
func CreateMap(mapType MapType, keySize, valueSize, maxEntries, flags, innerID uint32, path string) (int, error) {
	// This struct must be in sync with union bpf_attr's anonymous struct
	// used by the BPF_MAP_CREATE command
	uba := struct {
		mapType    uint32
		keySize    uint32
		valueSize  uint32
		maxEntries uint32
		mapFlags   uint32
		innerID    uint32
	}{
		uint32(mapType),
		keySize,
		valueSize,
		maxEntries,
		flags,
		innerID,
	}

	ret, _, err := unix.Syscall(
		unix.SYS_BPF,
		BPF_MAP_CREATE,
		uintptr(unsafe.Pointer(&uba)),
		unsafe.Sizeof(uba),
	)
	runtime.KeepAlive(&uba)
	if err != 0 {
		return 0, &os.PathError{
			Op:   "Unable to create map",
			Path: path,
			Err:  err,
		}
	}

	return int(ret), nil
}

// GetMapType determines whether the specified map type is supported by the
// kernel (as determined by bpftool feature checks), and if the map type is not
// supported, returns a more primitive map type that may be used to implement
// the map on older implementations. Otherwise, returns the specified map type.
func GetMapType(t MapType) MapType {
	pm := probes.NewProbeManager()
	supportedMapTypes := pm.GetMapTypes()
	switch t {
	case MapTypeLPMTrie, MapTypeLRUHash:
		if !supportedMapTypes.HaveLruHashMapType {
			return MapTypeHash
		}
	}
	return t
}

// This struct must be in sync with union bpf_attr's anonymous struct used by
// BPF_OBJ_*_ commands
type bpfAttrObjOp struct {
	pathname uint64
	fd       uint32
	pad0     [4]byte
}

// ObjGet reads the pathname and returns the map's fd read.
func ObjGet(pathname string) (int, error) {
	pathStr, err := unix.BytePtrFromString(pathname)
	if err != nil {
		return 0, fmt.Errorf("unable to convert pathname %q to byte pointer: %w", pathname, err)
	}
	bpfAttr := bpfAttrObjOp{
		pathname: uint64(uintptr(unsafe.Pointer(pathStr))),
	}

	fd, _, errno := unix.Syscall(
		unix.SYS_BPF,
		BPF_OBJ_GET,
		uintptr(unsafe.Pointer(&bpfAttr)),
		unsafe.Sizeof(bpfAttr),
	)
	runtime.KeepAlive(pathStr)
	runtime.KeepAlive(&bpfAttr)

	if fd == 0 || errno != 0 {
		return 0, &os.PathError{
			Op:   "Unable to get object",
			Err:  errno,
			Path: pathname,
		}
	}

	return int(fd), nil
}

// ObjPin stores the map's fd in pathname.
func ObjPin(fd int, pathname string) error {
	pathStr, err := unix.BytePtrFromString(pathname)
	if err != nil {
		return fmt.Errorf("Unable to convert pathname %q to byte pointer: %w", pathname, err)
	}
	bpfAttr := bpfAttrObjOp{
		pathname: uint64(uintptr(unsafe.Pointer(pathStr))),
		fd:       uint32(fd),
	}

	ret, _, errno := unix.Syscall(
		unix.SYS_BPF,
		BPF_OBJ_PIN,
		uintptr(unsafe.Pointer(&bpfAttr)),
		unsafe.Sizeof(bpfAttr),
	)
	runtime.KeepAlive(pathStr)
	runtime.KeepAlive(&bpfAttr)

	if ret != 0 || errno != 0 {
		return fmt.Errorf("unable to pin object with file descriptor %d to %s: %s", fd, pathname, errno)
	}

	return nil
}

func objCheck(fd int, path string, mapType MapType, keySize, valueSize, maxEntries, flags uint32) bool {
	info, err := GetMapInfo(os.Getpid(), fd)
	if err != nil {
		return false
	}

	mismatch := false
	if info.MapType != mapType ||
		info.KeySize != keySize ||
		info.ValueSize != valueSize ||
		info.MaxEntries != maxEntries ||
		info.Flags != flags {
		mismatch = true
	}

	if mismatch {
		if info.MapType == MapTypeProgArray {
			return false
		}

		klog.Warning("Removing map to allow for property upgrade (expect map data loss)")

		// Kernel still holds map reference count via attached prog.
		// Only exception is prog array, but that is already resolved
		// differently.
		os.Remove(path)
		return true
	}

	return false
}

// ObjClose closes the map's fd.
func ObjClose(fd int) error {
	if fd > 0 {
		return unix.Close(fd)
	}
	return nil
}

// GetPreAllocateMapFlags returns the map flags for map which use conditional
// pre-allocation.
func GetPreAllocateMapFlags(t MapType) uint32 {
	switch {
	case !t.allowsPreallocation():
		return BPF_F_NO_PREALLOC
	case t.requiresPreallocation():
		return 0
	}
	return atomic.LoadUint32(&preAllocateMapSetting)
}

// UnpinMapIfExists unpins the given map identified by name.
// If the map doesn't exist, returns success.
func UnpinMapIfExists(name string) error {
	path := MapPath(name)

	if _, err := os.Stat(path); err != nil {
		// Map doesn't exist
		return nil
	}

	return os.RemoveAll(path)
}

// This struct must be in sync with union bpf_attr's anonymous struct used by
// BPF_MAP_*_ELEM commands
type bpfAttrMapOpElem struct {
	mapFd uint32
	pad0  [4]byte
	key   uint64
	value uint64 // union: value or next_key
	flags uint64
}

// GetFirstKey fetches the first key in the map. If there are no keys in the
// map, io.EOF is returned.
func GetFirstKey(fd int, nextKey unsafe.Pointer) error {
	bpfAttr := bpfAttrMapOpElem{
		mapFd: uint32(fd),
		key:   0, // NULL -> Get first element
		value: uint64(uintptr(nextKey)),
	}

	ret := GetNextKeyFromPointers(fd, unsafe.Pointer(&bpfAttr), unsafe.Sizeof(bpfAttr))
	runtime.KeepAlive(nextKey)
	return ret
}

// GetNextKeyFromPointers stores, in nextKey, the next key after the key of the
// map in fd. When there are no more keys, io.EOF is returned.
func GetNextKeyFromPointers(fd int, structPtr unsafe.Pointer, sizeOfStruct uintptr) error {
	ret, _, err := unix.Syscall(
		unix.SYS_BPF,
		BPF_MAP_GET_NEXT_KEY,
		uintptr(structPtr),
		sizeOfStruct,
	)
	runtime.KeepAlive(structPtr)

	// BPF_MAP_GET_NEXT_KEY returns ENOENT when all keys have been iterated
	// translate that to io.EOF to signify there are no next keys
	if err == unix.ENOENT {
		return io.EOF
	}

	if ret != 0 || err != 0 {
		return fmt.Errorf("unable to get next key from map with file descriptor %d: %s", fd, err)
	}

	return nil
}

// DeleteElement deletes the map element with the given key.
func DeleteElement(fd int, key unsafe.Pointer) error {
	ret, err := deleteElement(fd, key)

	if ret != 0 || err != 0 {
		return fmt.Errorf("unable to delete element from map with file descriptor %d: %s", fd, err)
	}

	return nil
}

func deleteElement(fd int, key unsafe.Pointer) (uintptr, unix.Errno) {
	bpfAttr := bpfAttrMapOpElem{
		mapFd: uint32(fd),
		key:   uint64(uintptr(key)),
	}
	ret, _, err := unix.Syscall(
		unix.SYS_BPF,
		BPF_MAP_DELETE_ELEM,
		uintptr(unsafe.Pointer(&bpfAttr)),
		unsafe.Sizeof(bpfAttr),
	)
	runtime.KeepAlive(key)
	runtime.KeepAlive(&bpfAttr)

	return ret, err
}

// LookupElementFromPointers looks up for the map value stored in fd with the given key. The value
// is stored in the value unsafe.Pointer.
func LookupElementFromPointers(fd int, structPtr unsafe.Pointer, sizeOfStruct uintptr) error {
	ret, _, err := unix.Syscall(
		unix.SYS_BPF,
		BPF_MAP_LOOKUP_ELEM,
		uintptr(structPtr),
		sizeOfStruct,
	)
	runtime.KeepAlive(structPtr)

	if ret != 0 || err != 0 {
		return fmt.Errorf("Unable to lookup element in map with file descriptor %d: %s", fd, err)
	}

	return nil
}
