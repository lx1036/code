//go:build linux

package bpf

import (
	"fmt"
	"github.com/cilium/cilium/pkg/byteorder"
	"github.com/cilium/cilium/pkg/lock"
	"github.com/cilium/cilium/pkg/metrics"
	"os"
	"path"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf/binary"
)

type DumpParser func(key []byte, value []byte, mapKey MapKey, mapValue MapValue) (MapKey, MapValue, error)
type DumpCallback func(key MapKey, value MapValue)
type MapValidator func(path string) (bool, error)

type Map struct {
	MapInfo
	fd   int
	name string
	path string
	lock lock.RWMutex

	// inParallelMode is true when the Map is currently being run in
	// parallel and all modifications are performed on both maps until
	// EndParallelMode() is called.
	inParallelMode bool

	// cachedCommonName is the common portion of the name excluding any
	// endpoint ID
	cachedCommonName string

	// enableSync is true when synchronization retries have been enabled.
	enableSync bool

	// NonPersistent is true if the map does not contain persistent data
	// and should be removed on startup.
	NonPersistent bool

	// DumpParser is a function for parsing keys and values from BPF maps
	DumpParser DumpParser

	// withValueCache is true when map cache has been enabled
	withValueCache bool

	// cache as key/value entries when map cache is enabled or as key-only when
	// pressure metric is enabled
	cache map[string]*cacheEntry

	// errorResolverLastScheduled is the timestamp when the error resolver
	// was last scheduled
	errorResolverLastScheduled time.Time

	// outstandingErrors is the number of outsanding errors syncing with
	// the kernel
	outstandingErrors int

	// pressureGauge is a metric that tracks the pressure on this map
	pressureGauge *metrics.GaugeWithThreshold
}

type NewMapOpts struct {
	CheckValueSize bool // Enable mapValue and valueSize size check
}

func NewMap(name string, mapType MapType, mapKey MapKey, keySize int,
	mapValue MapValue, valueSize, maxEntries int, flags uint32, innerID uint32,
	dumpParser DumpParser) *Map {

	return NewMapWithOpts(name, mapType, mapKey, keySize, mapValue, valueSize,
		maxEntries, flags, innerID, dumpParser,
		&NewMapOpts{CheckValueSize: true})
}

func NewMapWithOpts(name string, mapType MapType, mapKey MapKey, keySize int,
	mapValue MapValue, valueSize, maxEntries int, flags uint32, innerID uint32,
	dumpParser DumpParser, opts *NewMapOpts) *Map {
	// 使用反射获取 mapKey 对象的 keySize，因为 mapKey 是接口 interface
	if size := reflect.TypeOf(mapKey).Elem().Size(); size != uintptr(keySize) {
		panic(fmt.Sprintf("Invalid %s map key size (%d != %d)", name, size, keySize))
	}

	if opts.CheckValueSize {
		if size := reflect.TypeOf(mapValue).Elem().Size(); size != uintptr(valueSize) {
			panic(fmt.Sprintf("Invalid %s map value size (%d != %d)", name, size, valueSize))
		}
	}

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
		DumpParser: dumpParser,
	}

	return m
}

func (m *Map) OpenOrCreate() (bool, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

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

// WithCache enables use of a cache. This will store all entries inserted from
// user space in a local cache (map) and will indicate the status of each
// individual entry.
func (m *Map) WithCache() *Map {
	if m.cache == nil {
		m.cache = map[string]*cacheEntry{}
	}
	m.withValueCache = true
	m.enableSync = true
	return m
}

// WithPressureMetricThreshold enables the tracking of a metric that measures
// the pressure of this map. This metric is only reported if over the
// threshold.
func (m *Map) WithPressureMetricThreshold(threshold float64) *Map {
	// When pressure metric is enabled, we keep track of map keys in cache
	if m.cache == nil {
		m.cache = map[string]*cacheEntry{}
	}

	m.pressureGauge = metrics.NewBPFMapPressureGauge(m.NonPrefixedName(), threshold)

	return m
}

// WithPressureMetric enables tracking and reporting of this map pressure with
// threshold 0.
func (m *Map) WithPressureMetric() *Map {
	return m.WithPressureMetricThreshold(0.0)
}

func (m *Map) NonPrefixedName() string {
	return strings.TrimPrefix(m.name, metrics.Namespace+"_")
}

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

type cacheEntry struct {
	Key   MapKey
	Value MapValue

	DesiredAction DesiredAction
	LastError     error
}

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

func (d DesiredAction) String() string {
	switch d {
	case OK:
		return "sync"
	case Insert:
		return "to-be-inserted"
	case Delete:
		return "to-be-deleted"
	default:
		return "unknown"
	}
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
