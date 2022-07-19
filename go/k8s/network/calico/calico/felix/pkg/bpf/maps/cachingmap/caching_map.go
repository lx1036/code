package cachingmap

import (
	"reflect"

	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf"
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf/maps"

	log "github.com/sirupsen/logrus"
)

// CachingMap INFO: 会把 service/endpoint bpf map 数据缓存到内存里进行计算，然后重刷回 bpf map
type CachingMap struct {
	// bpf map
	dataplaneMap bpf.Map
	params       maps.MapParameters

	// desiredStateOfDataplane stores the complete set of key/value pairs that we _want_ to
	// be in the dataplane.  Calling ApplyAllChanges attempts to bring the dataplane into
	// sync.
	//
	// For occupancy's sake we may want to drop this copy and instead maintain the invariant:
	// desiredStateOfDataplane = cacheOfDataplane - pendingDeletions + pendingUpdates.
	desiredStateOfDataplane *ByteArrayToByteArrayMap
	cacheOfDataplane        *ByteArrayToByteArrayMap
	pendingUpdates          *ByteArrayToByteArrayMap
	pendingDeletions        *ByteArrayToByteArrayMap
}

func New(mapParams maps.MapParameters, dataplaneMap bpf.Map) *CachingMap {
	cm := &CachingMap{
		params:                  mapParams,
		dataplaneMap:            dataplaneMap,
		desiredStateOfDataplane: NewByteArrayToByteArrayMap(mapParams.KeySize, mapParams.ValueSize),
	}

	return cm
}

func (c *CachingMap) LoadCacheFromDataplane() error {
	err := c.dataplaneMap.DumpWithCallback(func(k, v []byte) {
		c.cacheOfDataplane.Set(k, v)
	})
	if err != nil {
		c.clearCache()
		return err
	}

	c.recalculatePendingOperations()
	return nil
}

func (c *CachingMap) IterDataplaneCache(f func(k, v []byte)) {
	c.cacheOfDataplane.Iter(f)
}

func (c *CachingMap) clearCache() {
	c.cacheOfDataplane = nil
	c.pendingDeletions = nil
	c.pendingUpdates = nil
}

// ByteArrayToByteArrayMap INFO: 利用了 reflect.Value 有 interator map 函数，和存取数据机制
type ByteArrayToByteArrayMap struct {
	keySize   int
	valueSize int
	keyType   reflect.Type
	valueType reflect.Type

	m reflect.Value // map[[keySize]byte][valueSize]byte

	// key and value that we reuse when reading/writing the map.  Since the map uses value types (not
	// pointers), we can reuse the same key/value to read/write the map and the map will save the
	// actual key/value internally rather than sharing storage with our reflect.Value.
	key        reflect.Value
	value      reflect.Value
	keySlice   []byte // Slice backed by key
	valueSlice []byte // Slice backed by value
}

func NewByteArrayToByteArrayMap(keySize, valueSize int) *ByteArrayToByteArrayMap {
	// Effectively make(map[[keySize]byte][valueSize]byte)
	keyType := reflect.ArrayOf(keySize, reflect.TypeOf(byte(0)))
	valueType := reflect.ArrayOf(valueSize, reflect.TypeOf(byte(0)))
	mapType := reflect.MapOf(keyType, valueType)
	mapVal := reflect.MakeMap(mapType)
	key := reflect.New(keyType).Elem()
	value := reflect.New(valueType).Elem()
	return &ByteArrayToByteArrayMap{
		keySize:    keySize,
		valueSize:  valueSize,
		keyType:    keyType,
		valueType:  valueType,
		m:          mapVal,
		key:        key,
		value:      value,
		keySlice:   key.Slice(0, keySize).Interface().([]byte),
		valueSlice: value.Slice(0, valueSize).Interface().([]byte),
	}
}

func (b *ByteArrayToByteArrayMap) Set(k, v []byte) {
	if len(k) != b.keySize {
		log.Panic("ByteArrayToByteArrayMap.Set() called with incorrect key length")
	}
	if len(v) != b.valueSize {
		log.Panic("ByteArrayToByteArrayMap.Set() called with incorrect key length")
	}

	copy(b.keySlice, k)
	copy(b.valueSlice, v)
	b.m.SetMapIndex(b.key, b.value)
}

func (b *ByteArrayToByteArrayMap) Iter(f func(k, v []byte)) {
	iter := b.m.MapRange()
	// Since it's valid for a user to call Get/Set while we're iterating, make sure we have our own
	// values for key/value to avoid aliasing.
	key := reflect.New(b.keyType).Elem()
	val := reflect.New(b.valueType).Elem()
	keySlice := key.Slice(0, b.keySize).Interface().([]byte)
	valSlice := val.Slice(0, b.valueSize).Interface().([]byte)
	for iter.Next() {
		reflect.Copy(key, iter.Key())
		reflect.Copy(val, iter.Value())
		f(keySlice, valSlice)
	}
}
