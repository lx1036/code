package bpf

import (
	"fmt"
	"os"
	"path"
	"sync"
	"unsafe"
)

// MapType is an enumeration for valid BPF map types
type MapType int

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

func (m *Map) setPathIfUnset() error {
	if m.path == "" {
		if m.name == "" {
			return fmt.Errorf("either path or name must be set")
		}

		m.path = MapPath(m.name)
	}

	return nil
}
