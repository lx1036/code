package bpf

import (
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"runtime"
	"unsafe"
)

func GetMapFDByPin(filename string) (MapFD, error) {
	fd, err := ObjGet(filename)
	return MapFD(fd), err
}

type MapInfo struct {
	Type       int
	KeySize    int
	ValueSize  int
	MaxEntries int
}
type bpfAttrObjInfo struct {
	Fd      uint32
	InfoLen uint32
	Info    uint64
}
type bpfMapInfo struct {
	MapType    uint32
	MapID      uint32
	SizeKey    uint32
	SizeValue  uint32
	MaxEntries uint32
	Flags      uint32
}

// GetMapInfo INFO: https://github.com/cilium/cilium/blob/1.8.1/pkg/datapath/connector/ipvlan.go#L117-L136 根据 napID 获取 MapInfo
func GetMapInfo(fd MapFD) (*MapInfo, error) {
	info := bpfMapInfo{}
	bpfAttrInfo := bpfAttrObjInfo{
		Fd:      uint32(fd),
		InfoLen: uint32(unsafe.Sizeof(info)),
		Info:    uint64(uintptr(unsafe.Pointer(&info))),
	}
	bpfAttr2 := struct {
		info bpfAttrObjInfo
	}{
		info: bpfAttrInfo,
	}
	_, _, errno := unix.Syscall(
		unix.SYS_BPF,
		unix.BPF_OBJ_GET_INFO_BY_FD,
		uintptr(unsafe.Pointer(&bpfAttr2)),
		unsafe.Sizeof(bpfAttr2),
	)
	if errno != 0 {
		return nil, errno
	}

	return &MapInfo{
		Type:       int(info.MapType),
		KeySize:    int(info.SizeKey),
		ValueSize:  int(info.SizeValue),
		MaxEntries: int(info.MaxEntries),
	}, nil
}

func UpdateMapEntry(fd MapFD, k, v []byte) error {
	return UpdateElement(int(fd), unsafe.Pointer(&k), unsafe.Pointer(&v), unix.BPF_ANY)
}

// This struct must be in sync with union bpf_attr's anonymous struct used by
// BPF_OBJ_*_ commands
type bpfAttrObjOp struct {
	pathname uint64
	fd       uint32
	pad0     [4]byte
}

// ObjGet INFO: 根据 map filename 获取 bpf 虚拟文件系统 fd
// @see https://github.com/cilium/cilium/blob/v1.11.6/pkg/bpf/bpf_linux.go#L320-L355
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
		unix.BPF_OBJ_GET,
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

type bpfAttrFdFromId struct {
	ID     uint32
	NextID uint32
	Flags  uint32
}

// MapFdFromID INFO: 根据 mapID 获取 bpf 虚拟文件系统 fd
// @see https://github.com/cilium/cilium/blob/v1.11.6/pkg/bpf/bpf_linux.go#L363-L389
func MapFdFromID(id int) (int, error) {
	bpfAttr := bpfAttrFdFromId{
		ID: uint32(id),
	}
	fd, _, err := unix.Syscall(
		unix.SYS_BPF,
		unix.BPF_MAP_GET_FD_BY_ID,
		uintptr(unsafe.Pointer(&bpfAttr)),
		unsafe.Sizeof(bpfAttr),
	)
	runtime.KeepAlive(&bpfAttr)

	if fd == 0 || err != 0 {
		return 0, fmt.Errorf("Unable to get object fd from id %d: %s", id, err)
	}

	return int(fd), nil
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

// UpdateElement INFO: https://github.com/cilium/cilium/blob/v1.11.6/pkg/bpf/bpf_linux.go#L121-L139
func UpdateElement(fd int, key, value unsafe.Pointer, flags uint64) error {
	bpfAttr := bpfAttrMapOpElem{
		mapFd: uint32(fd),
		key:   uint64(uintptr(key)),
		value: uint64(uintptr(value)),
		flags: uint64(flags),
	}

	ret := UpdateElementFromPointers(fd, unsafe.Pointer(&bpfAttr), unsafe.Sizeof(bpfAttr))
	runtime.KeepAlive(key)
	runtime.KeepAlive(value)
	return ret
}

// UpdateElementFromPointers updates the map in fd with the given value in the given key.
func UpdateElementFromPointers(fd int, structPtr unsafe.Pointer, sizeOfStruct uintptr) error {
	ret, _, err := unix.Syscall(
		unix.SYS_BPF,
		unix.BPF_MAP_UPDATE_ELEM,
		uintptr(structPtr),
		sizeOfStruct,
	)
	runtime.KeepAlive(structPtr)
	if ret != 0 || err != 0 {
		return fmt.Errorf("Unable to update element for map with file descriptor %d: %s", fd, err)
	}

	return nil
}

func GetMapEntry(fd MapFD, key []byte, valueSize int) ([]byte, error) {
	value := make([]byte, valueSize)
	err := LookupElement(int(fd), unsafe.Pointer(&key), unsafe.Pointer(&value))
	if err != nil {
		return nil, err
	}

	return value, nil
}

// LookupElement INFO: 从 map fd 中查找 key 对应的 value
func LookupElement(fd int, key, value unsafe.Pointer) error {
	uba := bpfAttrMapOpElem{
		mapFd: uint32(fd),
		key:   uint64(uintptr(key)),
		value: uint64(uintptr(value)),
	}

	ret := LookupElementFromPointers(fd, unsafe.Pointer(&uba), unsafe.Sizeof(uba))
	runtime.KeepAlive(key)
	runtime.KeepAlive(value)
	return ret
}
func LookupElementFromPointers(fd int, structPtr unsafe.Pointer, sizeOfStruct uintptr) error {
	ret, _, err := unix.Syscall(
		unix.SYS_BPF,
		unix.BPF_MAP_LOOKUP_ELEM,
		uintptr(structPtr),
		sizeOfStruct,
	)
	runtime.KeepAlive(structPtr)

	if ret != 0 || err != 0 {
		return fmt.Errorf("Unable to lookup element in map with file descriptor %d: %s", fd, err)
	}

	return nil
}

func DeleteMapEntry(mapFD MapFD, k []byte, valueSize int) error {
	return DeleteElement(int(mapFD), unsafe.Pointer(&k))
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
		unix.BPF_MAP_DELETE_ELEM,
		uintptr(unsafe.Pointer(&bpfAttr)),
		unsafe.Sizeof(bpfAttr),
	)
	runtime.KeepAlive(key)
	runtime.KeepAlive(&bpfAttr)

	return ret, err
}

// GetNextKeyFromPointers stores, in nextKey, the next key after the key of the
// map in fd. When there are no more keys, io.EOF is returned.
func GetNextKeyFromPointers(fd int, structPtr unsafe.Pointer, sizeOfStruct uintptr) error {
	ret, _, err := unix.Syscall(
		unix.SYS_BPF,
		unix.BPF_MAP_GET_NEXT_KEY,
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
