package bpf

import (
	"bufio"
	"fmt"
	"github.com/cilium/cilium/pkg/byteorder"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"path"
	"sync"
	"unsafe"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/binary"
)

// MapType is an enumeration for valid BPF map types
type MapType int

func (t MapType) String() string {
	switch t {
	case MapTypeHash:
		return "Hash"
	case MapTypeArray:
		return "Array"
	case MapTypeProgArray:
		return "Program array"
	case MapTypePerfEventArray:
		return "Event array"
	case MapTypePerCPUHash:
		return "Per-CPU hash"
	case MapTypePerCPUArray:
		return "Per-CPU array"
	case MapTypeStackTrace:
		return "Stack trace"
	case MapTypeCgroupArray:
		return "Cgroup array"
	case MapTypeLRUHash:
		return "LRU hash"
	case MapTypeLRUPerCPUHash:
		return "LRU per-CPU hash"
	case MapTypeLPMTrie:
		return "Longest prefix match trie"
	case MapTypeArrayOfMaps:
		return "Array of maps"
	case MapTypeHashOfMaps:
		return "Hash of maps"
	case MapTypeDevMap:
		return "Device Map"
	case MapTypeSockMap:
		return "Socket Map"
	case MapTypeCPUMap:
		return "CPU Redirect Map"
	case MapTypeSockHash:
		return "Socket Hash"
	}

	return "Unknown"
}

func (t MapType) allowsPreallocation() bool {
	return t != MapTypeLPMTrie
}

func (t MapType) requiresPreallocation() bool {
	switch t {
	case MapTypeHash, MapTypePerCPUHash, MapTypeLPMTrie, MapTypeHashOfMaps:
		return false
	}
	return true
}

// This enumeration must be in sync with enum bpf_map_type in <linux/bpf.h>
const (
	MapTypeUnspec MapType = iota
	MapTypeHash
	MapTypeArray
	MapTypeProgArray
	MapTypePerfEventArray
	MapTypePerCPUHash
	MapTypePerCPUArray
	MapTypeStackTrace
	MapTypeCgroupArray
	MapTypeLRUHash
	MapTypeLRUPerCPUHash
	MapTypeLPMTrie
	MapTypeArrayOfMaps
	MapTypeHashOfMaps
	MapTypeDevMap
	MapTypeSockMap
	MapTypeCPUMap
	MapTypeXSKMap
	MapTypeSockHash
	// MapTypeMaximum is the maximum supported known map type.
	MapTypeMaximum
)

type MapKey interface {
	fmt.Stringer

	// Returns pointer to start of key
	GetKeyPtr() unsafe.Pointer

	// Allocates a new value matching the key type
	NewValue() MapValue

	// DeepCopyMapKey returns a deep copy of the map key
	DeepCopyMapKey() MapKey
}

type MapValue interface {
	fmt.Stringer

	// Returns pointer to start of value
	GetValuePtr() unsafe.Pointer

	// DeepCopyMapValue returns a deep copy of the map value
	DeepCopyMapValue() MapValue
}

type MapInfo struct {
	MapType  MapType
	MapKey   MapKey
	KeySize  uint32
	MapValue MapValue
	// ReadValueSize is the value size that is used to read from the BPF maps
	// this value and the ValueSize values can be different for MapTypePerCPUHash.
	ReadValueSize uint32
	ValueSize     uint32
	MaxEntries    uint32
	Flags         uint32
	InnerID       uint32
	OwnerProgType ProgType
}

/*
GetMapInfo 获取
cat /proc/17178/fdinfo/17
pos:	0
flags:	02000002
mnt_id:	13
map_type:	11
key_size:	12
value_size:	1
max_entries:	65536
map_flags:	0x1
memlock:	3215360
map_id:	5089
*/
func GetMapInfo(pid int, fd int) (*MapInfo, error) {
	fdinfoFile := fmt.Sprintf("/proc/%d/fdinfo/%d", pid, fd)
	file, err := os.Open(fdinfoFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info := &MapInfo{}

	// INFO: 这里可以借鉴，如何读取内容少量格式固定的文件
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		var value int

		line := scanner.Text()
		if n, err := fmt.Sscanf(line, "map_type:\t%d", &value); n == 1 && err == nil {
			info.MapType = MapType(value)
		} else if n, err := fmt.Sscanf(line, "key_size:\t%d", &value); n == 1 && err == nil {
			info.KeySize = uint32(value)
		} else if n, err := fmt.Sscanf(line, "value_size:\t%d", &value); n == 1 && err == nil {
			info.ValueSize = uint32(value)
			info.ReadValueSize = uint32(value)
		} else if n, err := fmt.Sscanf(line, "max_entries:\t%d", &value); n == 1 && err == nil {
			info.MaxEntries = uint32(value)
		} else if n, err := fmt.Sscanf(line, "map_flags:\t0x%x", &value); n == 1 && err == nil {
			info.Flags = uint32(value)
		} else if n, err := fmt.Sscanf(line, "owner_prog_type:\t%d", &value); n == 1 && err == nil {
			info.OwnerProgType = ProgType(value)
		}
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	return info, nil
}

type DumpParser func(key []byte, value []byte, mapKey MapKey, mapValue MapValue) (MapKey, MapValue, error)
type DumpCallback func(key MapKey, value MapValue)

// DesiredAction is the action to be performed on the BPF map
type DesiredAction int

const (
	// OK indicates that to further action is required and the entry is in
	// sync
	OK DesiredAction = iota

	// Insert indicates that the entry needs to be created or updated
	Insert

	// Delete indicates that the entry needs to be deleted
	Delete
)

type cacheEntry struct {
	Key   MapKey
	Value MapValue

	DesiredAction DesiredAction
	LastError     error
}

type Map struct {
	sync.RWMutex

	MapInfo

	fd   int
	name string
	path string

	cache map[string]*cacheEntry
	// enableSync is true when synchronization retries have been enabled.
	enableSync bool

	// NonPersistent is true if the map does not contain persistent data
	// and should be removed on startup.
	NonPersistent bool

	// DumpParser is a function for parsing keys and values from BPF maps
	dumpParser DumpParser
}

// NewMap creates a new Map instance - object representing a BPF map
func NewMap(name string, mapType MapType, mapKey MapKey, keySize int, mapValue MapValue, valueSize,
	maxEntries int, flags uint32, innerID uint32, dumpParser DumpParser) *Map {
	m := &Map{
		MapInfo: MapInfo{
			MapType:       mapType,
			MapKey:        mapKey,
			KeySize:       uint32(keySize),
			MapValue:      mapValue,
			ReadValueSize: uint32(valueSize),
			ValueSize:     uint32(valueSize),
			MaxEntries:    uint32(maxEntries),
			Flags:         flags,
			InnerID:       innerID,
			OwnerProgType: ProgTypeUnspec,
		},
		name:       path.Base(name),
		dumpParser: dumpParser,
	}
	return m
}

// WithCache enables use of a cache. This will store all entries inserted from
// user space in a local cache (map) and will indicate the status of each
// individual entry.
func (m *Map) WithCache() *Map {
	m.cache = map[string]*cacheEntry{}
	m.enableSync = true
	return m
}

// OpenOrCreate attempts to open the Map, or if it does not yet exist, create
// the Map. If the existing map's attributes such as map type, key/value size,
// capacity, etc. do not match the Map's attributes, then the map will be
// deleted and reopened without any attempt to retain its previous contents.
// If the map is marked as non-persistent, it will always be recreated.
//
// If the map type is MapTypeLRUHash or MapTypeLPMTrie and the kernel lacks
// support for this map type, then the map will be opened as MapTypeHash
// instead. Note that the BPF code that interacts with this map *MUST* be
// structured in such a way that the map is declared as the same type based on
// the same probe logic (eg HAVE_LRU_HASH_MAP_TYPE, HAVE_LPM_TRIE_MAP_TYPE).
//
// For code that uses an LPMTrie, the BPF code must also use macros to retain
// the "longest prefix match" behaviour on top of the hash maps, for example
// via LPM_LOOKUP_FN() (see bpf/lib/maps.h).
//
// Returns whether the map was deleted and recreated, or an optional error.
func (m *Map) OpenOrCreate() (bool, error) {
	m.Lock()
	defer m.Unlock()

	return m.openOrCreate(true)
}

func (m *Map) openOrCreate(pin bool) (bool, error) {
	if m.fd != 0 {
		return false, nil
	}

	if err := m.setPathIfUnset(); err != nil {
		return false, err
	}

	// If the map represents non-persistent data, always remove the map
	// before opening or creating.
	if m.NonPersistent {
		os.Remove(m.path)
	}

	mapType := GetMapType(m.MapType)
	flags := m.Flags | GetPreAllocateMapFlags(mapType)
	fd, isNew, err := OpenOrCreateMap(m.path, mapType, m.KeySize, m.ValueSize, m.MaxEntries, flags, m.InnerID, pin)
	if err != nil {
		return false, err
	}

	registerMap(m.path, m)

	m.fd = fd
	m.MapType = mapType
	m.Flags = flags
	return isNew, nil
}

func (m *Map) Open() error {
	m.Lock()
	defer m.Unlock()

	return m.open()
}

func (m *Map) open() error {
	if m.fd != 0 {
		return nil
	}

	if err := m.setPathIfUnset(); err != nil {
		return err
	}

	fd, err := ObjGet(m.path)
	if err != nil {
		return err
	}

	registerMap(m.path, m)

	m.fd = fd
	m.MapType = GetMapType(m.MapType)
	return nil
}

func (m *Map) setPathIfUnset() error {
	if m.path == "" {
		if m.name == "" {
			return fmt.Errorf("either path or name must be set")
		}

		m.path = MapPath(m.name)
	}

	return nil
}

// Create is similar to OpenOrCreate, but closes the map after creating or
// opening it.
func (m *Map) Create() (bool, error) {
	isNew, err := m.OpenOrCreate()
	if err != nil {
		return isNew, err
	}
	return isNew, m.Close()
}

// DumpWithCallback iterates over the Map and calls the given callback
// function on each iteration. That callback function is receiving the
// actual key and value. The callback function should consider creating a
// deepcopy of the key and value on between each iterations to avoid memory
// corruption.
func (m *Map) DumpWithCallback(cb DumpCallback) error {
	if err := m.Open(); err != nil {
		return err
	}

	m.RLock()
	defer m.RUnlock()

	key := make([]byte, m.KeySize)
	nextKey := make([]byte, m.KeySize)
	value := make([]byte, m.ReadValueSize)

	if err := GetFirstKey(m.fd, unsafe.Pointer(&nextKey[0])); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	bpfCurrentKey := bpfAttrMapOpElem{
		mapFd: uint32(m.fd),
		key:   uint64(uintptr(unsafe.Pointer(&key[0]))),
		value: uint64(uintptr(unsafe.Pointer(&nextKey[0]))),
	}
	bpfCurrentKeyPtr := unsafe.Pointer(&bpfCurrentKey)
	bpfCurrentKeySize := unsafe.Sizeof(bpfCurrentKey)

	bpfNextKey := bpfAttrMapOpElem{
		mapFd: uint32(m.fd),
		key:   uint64(uintptr(unsafe.Pointer(&nextKey[0]))),
		value: uint64(uintptr(unsafe.Pointer(&value[0]))),
	}
	bpfNextKeyPtr := unsafe.Pointer(&bpfNextKey)
	bpfNextKeySize := unsafe.Sizeof(bpfNextKey)

	mk := m.MapKey.DeepCopyMapKey()
	mv := m.MapValue.DeepCopyMapValue()

	for {
		err := LookupElementFromPointers(m.fd, bpfNextKeyPtr, bpfNextKeySize)
		if err != nil {
			return err
		}

		mk, mv, err = m.dumpParser(nextKey, value, mk, mv)
		if err != nil {
			return err
		}

		if cb != nil {
			cb(mk, mv)
		}

		copy(key, nextKey)

		if err := GetNextKeyFromPointers(m.fd, bpfCurrentKeyPtr, bpfCurrentKeySize); err != nil {
			if err == io.EOF { // end of map, we're done iterating
				return nil
			}
			return err
		}
	}
}

// DeleteAll deletes all entries of a map by traversing the map and deleting individual
// entries. Note that if entries are added while the taversal is in progress,
// such entries may survive the deletion process.
func (m *Map) DeleteAll() error {
	m.Lock()
	defer m.Unlock()

	nextKey := make([]byte, m.KeySize)

	if m.cache != nil {
		// Mark all entries for deletion, upon successful deletion,
		// entries will be removed or the LastError will be updated
		for _, entry := range m.cache {
			entry.DesiredAction = Delete
			entry.LastError = fmt.Errorf("deletion pending")
		}
	}

	if err := m.open(); err != nil {
		return err
	}

	mk := m.MapKey.DeepCopyMapKey()
	mv := m.MapValue.DeepCopyMapValue()

	for {
		if err := GetFirstKey(m.fd, unsafe.Pointer(&nextKey[0])); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		err := DeleteElement(m.fd, unsafe.Pointer(&nextKey[0]))

		mk, _, err2 := m.dumpParser(nextKey, []byte{}, mk, mv)
		if err2 == nil {
			m.deleteCacheEntry(mk, err)
		} else {
			log.WithError(err2).Warningf("Unable to correlate iteration key %v with cache entry. Inconsistent cache.", nextKey)
		}

		if err != nil {
			return err
		}
	}
}

// Delete deletes the map entry corresponding to the given key.
func (m *Map) Delete(key MapKey) error {
	m.Lock()
	defer m.Unlock()

	var err error
	defer m.deleteCacheEntry(key, err)

	if err = m.open(); err != nil {
		return err
	}

	_, errno := deleteElement(m.fd, key.GetKeyPtr())
	if errno != 0 {
		err = fmt.Errorf("unable to delete element %s from map %s: %w", key, m.name, errno)
	}
	return err
}

// Caller must hold m.lock for writing
func (m *Map) deleteCacheEntry(key MapKey, err error) {
	if m.cache == nil {
		return
	}

	k := key.String()
	if err == nil {
		delete(m.cache, k)
	} else {
		entry, ok := m.cache[k]
		if !ok {
			m.cache[k] = &cacheEntry{
				Key: key,
			}
			entry = m.cache[k]
		}

		entry.DesiredAction = Delete
		entry.LastError = err
		//m.scheduleErrorResolver()
	}
}

func (m *Map) Close() error {
	m.Lock()
	defer m.Unlock()

	if m.enableSync {
		//mapControllers.RemoveController(m.controllerName())
	}

	if m.fd != 0 {
		unix.Close(m.fd)
		m.fd = 0
	}

	unregisterMap(m.path, m)

	return nil
}

// ConvertKeyValue converts key and value from bytes to given Golang struct pointers.
func ConvertKeyValue(bKey []byte, bValue []byte, key MapKey, value MapValue) (MapKey, MapValue, error) {
	if len(bKey) > 0 {
		if err := binary.Read(bKey, byteorder.Native, key); err != nil {
			return nil, nil, fmt.Errorf("Unable to convert key: %s", err)
		}
	}

	if len(bValue) > 0 {
		if err := binary.Read(bValue, byteorder.Native, value); err != nil {
			return nil, nil, fmt.Errorf("Unable to convert value: %s", err)
		}
	}

	return key, value, nil
}
