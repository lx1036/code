//go:build linux

package bpf

import (
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf/binary"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/byteorder"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/controller"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/metrics"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"

	"golang.org/x/sys/unix"
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

func (m *Map) GetFd() int {
	return m.fd
}

// Open map path -> map fd
func (m *Map) Open() error {
	m.lock.Lock()
	defer m.lock.Unlock()

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

// Create is similar to OpenOrCreate, but closes the map after creating or
// opening it.
func (m *Map) Create() (bool, error) {
	isNew, err := m.OpenOrCreate()
	if err != nil {
		return isNew, err
	}
	return isNew, m.Close()
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

func (m *Map) DumpWithCallback(cb DumpCallback) error {
	if err := m.Open(); err != nil {
		return err
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	key := make([]byte, m.KeySize)
	nextKey := make([]byte, m.KeySize)
	value := make([]byte, m.ReadValueSize)

	if err := GetFirstKey(m.fd, unsafe.Pointer(&nextKey[0])); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	mk := m.MapKey.DeepCopyMapKey()
	mv := m.MapValue.DeepCopyMapValue()

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

	for {
		err := LookupElementFromPointers(m.fd, bpfNextKeyPtr, bpfNextKeySize)
		if err != nil {
			return err
		}

		mk, mv, err = m.DumpParser(nextKey, value, mk, mv)
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

func (m *Map) Update(key MapKey, value MapValue) error {
	var err error

	m.lock.Lock()
	defer m.lock.Unlock()

	defer func() {
		if m.cache == nil {
			return
		}

		if m.withValueCache {
			desiredAction := OK
			if err != nil {
				desiredAction = Insert
				m.scheduleErrorResolver()
			}

			m.cache[key.String()] = &cacheEntry{
				Key:           key,
				Value:         value,
				DesiredAction: desiredAction,
				LastError:     err,
			}
			m.updatePressureMetric()
		} else if err == nil {
			m.cache[key.String()] = nil
			m.updatePressureMetric()
		}
	}()

	if err = m.open(); err != nil {
		return err
	}

	err = UpdateElement(m.fd, m.name, key.GetKeyPtr(), value.GetValuePtr(), 0)
	if option.Config.MetricsConfig.BPFMapOps {
		metrics.BPFMapOps.WithLabelValues(m.commonName(), metricOpUpdate, metrics.Error2Outcome(err)).Inc()
	}

	return err
}

func (m *Map) Lookup(key MapKey) (MapValue, error) {
	if err := m.Open(); err != nil {
		return nil, err
	}

	value := key.NewValue()

	m.lock.RLock()
	defer m.lock.RUnlock()

	err := LookupElement(m.fd, key.GetKeyPtr(), value.GetValuePtr())
	if err != nil {
		return nil, err
	}
	return value, nil
}

// Delete deletes the map entry corresponding to the given key.
func (m *Map) Delete(key MapKey) error {
	_, err := m.deleteMapEntry(key, false)
	return err
}

// SilentDelete deletes the map entry corresponding to the given key.
// If a map entry is not found this returns (true, nil).
func (m *Map) SilentDelete(key MapKey) (deleted bool, err error) {
	return m.deleteMapEntry(key, true)
}

// deleteMapEntry deletes the map entry corresponding to the given key.
// If ignoreMissing is set to true and the entry is not found, then
// the error metric is not incremented for missing entries and nil error is returned.
func (m *Map) deleteMapEntry(key MapKey, ignoreMissing bool) (deleted bool, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	defer func() {
		m.deleteCacheEntry(key, err)
		if err != nil {
			m.updatePressureMetric()
		}
	}()

	if err = m.open(); err != nil {
		return false, err
	}

	_, errno := deleteElement(m.fd, key.GetKeyPtr())
	deleted = errno == 0
	// Error handling is skipped in the case ignoreMissing is set and the
	// error is ENOENT. This removes false positives in the delete metrics
	// and skips the deferred cleanup of non-existing entries. This situation
	// occurs at least in the context of cleanup of NAT mappings from CT GC.
	handleError := errno != unix.ENOENT || !ignoreMissing
	if option.Config.MetricsConfig.BPFMapOps && handleError {
		metrics.BPFMapOps.WithLabelValues(m.commonName(), metricOpDelete, metrics.Errno2Outcome(errno)).Inc()
	}
	if errno != 0 && handleError {
		err = fmt.Errorf("unable to delete element %s from map %s: %w", key, m.name, errno)
	}

	return
}

// deleteCacheEntry evaluates the specified error, if nil the map key is
// removed from the cache to indicate successful deletion. If non-nil, the map
// key entry in the cache is updated to indicate deletion failure with the
// specified error.
//
// Caller must hold m.lock for writing
func (m *Map) deleteCacheEntry(key MapKey, err error) {
	if m.cache == nil {
		return
	}

	k := key.String()
	if err == nil {
		delete(m.cache, k)
	} else if !m.withValueCache {
		return
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
		m.scheduleErrorResolver()
	}
}

// scheduleErrorResolver schedules a periodic resolver controller that scans
// all BPF map caches for unresolved errors and attempts to resolve them. On
// error of resolution, the controller is-rescheduled in an expedited manner
// with an exponential back-off.
//
// m.lock must be held for writing
func (m *Map) scheduleErrorResolver() {
	m.outstandingErrors++

	if time.Since(m.errorResolverLastScheduled) <= errorResolverSchedulerMinInterval {
		return
	}

	m.errorResolverLastScheduled = time.Now()
	go func() {
		time.Sleep(errorResolverSchedulerDelay)
		mapControllers.UpdateController(m.controllerName(),
			controller.ControllerParams{
				DoFunc:      m.resolveErrors,
				RunInterval: errorResolverSchedulerMinInterval,
			},
		)
	}()
}

func (m *Map) updatePressureMetric() {
	if m.pressureGauge == nil {
		return
	}

	// Do a lazy check of MetricsConfig as it is not available at map static
	// initialization.
	if !option.Config.MetricsConfig.BPFMapPressure {
		if !m.withValueCache {
			m.cache = nil
		}
		m.pressureGauge = nil
		return
	}

	pvalue := float64(len(m.cache)) / float64(m.MaxEntries)
	m.pressureGauge.Set(pvalue)
}

// WithNonPersistent turns the map non-persistent and returns the map
func (m *Map) WithNonPersistent() *Map {
	m.NonPersistent = true
	return m
}

func (m *Map) commonName() string {
	if m.cachedCommonName != "" {
		return m.cachedCommonName
	}

	m.cachedCommonName = extractCommonName(m.name)
	return m.cachedCommonName
}

// Reopen attempts to close and re-open the received map.
func (m *Map) Reopen() error {
	m.Close()
	return m.Open()
}

func (m *Map) controllerName() string {
	return fmt.Sprintf("bpf-map-sync-%s", m.name)
}

func (m *Map) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.enableSync {
		mapControllers.RemoveController(m.controllerName())
	}

	if m.fd != 0 {
		unix.Close(m.fd)
		m.fd = 0
	}

	unregisterMap(m.path, m)

	return nil
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
