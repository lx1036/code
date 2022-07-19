package bpf

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"io"
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf/maps"
	"os"
	"os/exec"
	"strings"
	"unsafe"

	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf/maps/arp"
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf/maps/conntrack"
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf/maps/ipsets"
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf/maps/nat"
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf/maps/routes"

	"k8s.io/klog/v2"
)

type IteratorAction string

const (
	BPFBasePath = "/sys/fs/bpf/tc/globals"
)

const (
	IterNone   IteratorAction = ""
	IterDelete IteratorAction = "delete"
)

type IterCallback func(k, v []byte)

type MapContext struct {
	RepinningEnabled bool
	IpsetsMap        *Map
	StateMap         *Map
	ArpMap           *Map
	FailsafesMap     *Map
	FrontendMap      *Map
	BackendMap       *Map
	AffinityMap      *Map
	RouteMap         *Map
	CtMap            *Map
	SrMsgMap         *Map
	CtNatsMap        *Map
	MapSizes         map[string]uint32
}

func (c *MapContext) NewPinnedMap(params maps.MapParameters) *Map {
	if len(params.VersionedName()) >= unix.BPF_OBJ_NAME_LEN {
		log.WithField("name", params.Name).Panic("Bug: BPF map name too long")
	}
	if val, ok := c.MapSizes[params.VersionedName()]; ok {
		params.MaxEntries = int(val)
	}

	m := &Map{
		context:       c,
		MapParameters: params,
		perCPU:        strings.Contains(params.Type, "percpu"),
	}
	return m
}

func CreateBPFMapContext(ipsetsMapSize, natFEMapSize, natBEMapSize, natAffMapSize, routeMapSize,
	ctMapSize int, repinEnabled bool) *MapContext {
	bpfMapContext := &MapContext{
		RepinningEnabled: repinEnabled,
		MapSizes:         map[string]uint32{},
	}
	bpfMapContext.MapSizes[ipsets.MapParameters.VersionedName()] = uint32(ipsetsMapSize)
	bpfMapContext.MapSizes[nat.FrontendMapParameters.VersionedName()] = uint32(natFEMapSize)
	bpfMapContext.MapSizes[nat.BackendMapParameters.VersionedName()] = uint32(natBEMapSize)
	bpfMapContext.MapSizes[nat.AffinityMapParameters.VersionedName()] = uint32(natAffMapSize)
	bpfMapContext.MapSizes[routes.MapParameters.VersionedName()] = uint32(routeMapSize)
	bpfMapContext.MapSizes[conntrack.MapParams.VersionedName()] = uint32(ctMapSize)

	bpfMapContext.MapSizes[arp.MapParams.VersionedName()] = uint32(arp.MapParams.MaxEntries)
	bpfMapContext.MapSizes[nat.SendRecvMsgMapParameters.VersionedName()] = uint32(nat.SendRecvMsgMapParameters.MaxEntries)
	bpfMapContext.MapSizes[nat.CTNATsMapParameters.VersionedName()] = uint32(nat.CTNATsMapParameters.MaxEntries)

	return bpfMapContext
}

func CreateBPFMaps(mc *MapContext) error {
	var bpfMaps []*Map

	mc.IpsetsMap = ipsets.Map(mc)
	bpfMaps = append(bpfMaps, mc.IpsetsMap)

	mc.ArpMap = arp.Map(mc)
	bpfMaps = append(bpfMaps, mc.ArpMap)

	mc.FrontendMap = nat.FrontendMap(mc)
	bpfMaps = append(bpfMaps, mc.FrontendMap)

	mc.BackendMap = nat.BackendMap(mc)
	bpfMaps = append(bpfMaps, mc.BackendMap)

	mc.AffinityMap = nat.AffinityMap(mc)
	bpfMaps = append(bpfMaps, mc.AffinityMap)

	mc.RouteMap = routes.Map(mc)
	bpfMaps = append(bpfMaps, mc.RouteMap)

	mc.CtMap = conntrack.Map(mc)
	bpfMaps = append(bpfMaps, mc.CtMap)

	mc.SrMsgMap = nat.SendRecvMsgMap(mc)
	bpfMaps = append(bpfMaps, mc.SrMsgMap)

	mc.CtNatsMap = nat.AllNATsMsgMap(mc)
	bpfMaps = append(bpfMaps, mc.CtNatsMap)

	for _, bpfMap := range bpfMaps {
		err := bpfMap.EnsureExists()
		if err != nil {
			return fmt.Errorf("failed to create %s map, err=%w", bpfMap.GetName(), err)
		}
	}
	return nil
}

type Map struct {
	context *MapContext
	maps.MapParameters

	fdLoaded bool
	fd       MapFD
	oldfd    MapFD
	perCPU   bool
	oldSize  int
}

func (m *Map) GetName() string {
	return m.VersionedName()
}

// EnsureExists INFO: 使用 bpftool 创建 bpf map
func (m *Map) EnsureExists() error {
	if m.fdLoaded {
		return nil
	}

	// 这里有个逻辑：防止 bpf map MaxEntries 发生了变化
	if err := m.Open(); err == nil {
		// Get the existing map info
		mapInfo, err := GetMapInfo(m.fd)
		if err != nil {
			return fmt.Errorf("error getting map info of the pinned map %w", err)
		}
		if m.MaxEntries == mapInfo.MaxEntries {
			return nil
		}
	}

	// bpftool map create FILE type TYPE key KEY_SIZE value VALUE_SIZE entries MAX_ENTRIES name NAME [flags FLAGS] [dev NAME]
	klog.Infof(fmt.Sprintf("Map %s didn't exist, creating it", m.VersionedName()))
	cmd := exec.Command("bpftool", "map", "create", m.VersionedFilename(),
		"type", m.Type,
		"key", fmt.Sprint(m.KeySize),
		"value", fmt.Sprint(m.ValueSize),
		"entries", fmt.Sprint(m.MaxEntries),
		"name", m.VersionedName(),
		"flags", fmt.Sprint(m.Flags),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithField("out", string(out)).Error("Failed to run bpftool")
		return err
	}
	m.fd, err = GetMapFDByPin(m.VersionedFilename())
	if err == nil {
		m.fdLoaded = true
		log.WithField("fd", m.fd).WithField("name", m.VersionedFilename()).
			Info("Loaded map file descriptor.")
	}

	return err
}

func (m *Map) Open() error {
	if m.fdLoaded {
		return nil
	}

	_, err := MaybeMountBPFfs()
	if err != nil {
		log.WithError(err).Error("Failed to mount bpffs")
		return err
	}

	// FIXME hard-coded dir
	if _, err := os.Stat(BPFBasePath); os.IsNotExist(err) {
		err = os.MkdirAll(BPFBasePath, 0700)
		if err != nil {
			log.WithError(err).Error("Failed create dir")
			return err
		}
	}

	_, err = os.Stat(m.VersionedFilename())
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		log.Debug("Map file didn't exist")
		if m.context.RepinningEnabled {
			log.WithField("name", m.Name).Info("Looking for map by name (to repin it)")
			err = RepinMap(m.VersionedName(), m.VersionedFilename())
			if err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}

	if err == nil {
		log.Debug("Map file already exists, trying to open it")
		m.fd, err = GetMapFDByPin(m.VersionedFilename())
		if err == nil {
			m.fdLoaded = true
			log.WithField("fd", m.fd).WithField("name", m.VersionedFilename()).
				Info("Loaded map file descriptor.")
			return nil
		}
		return err
	}

	return err
}

func (m *Map) MapFD() MapFD {
	if !m.fdLoaded {
		log.WithField("map", *m).Panic("MapFD() called without first calling EnsureExists()")
	}

	return m.fd
}

func (m *Map) Path() string {
	return m.VersionedFilename()
}

func (m *Map) CopyDeltaFromOldMap() error {
	//TODO implement me
	panic("implement me")
}

// DumpWithCallback 迭代每一个 key-value
func (m *Map) DumpWithCallback(cb IterCallback) error {
	key := make([]byte, m.KeySize)
	nextKey := make([]byte, m.KeySize)
	value := make([]byte, m.ValueSize)

	if err := GetFirstKey(int(m.fd), unsafe.Pointer(&nextKey[0])); err != nil {
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

	for {
		err := LookupElementFromPointers(int(m.fd), bpfNextKeyPtr, bpfNextKeySize)
		if err != nil {
			return err
		}

		if cb != nil {
			cb(nextKey, value)
		}

		copy(key, nextKey)

		if err := GetNextKeyFromPointers(int(m.fd), bpfCurrentKeyPtr, bpfCurrentKeySize); err != nil {
			if err == io.EOF { // end of map, we're done iterating
				return nil
			}
			return err
		}
	}
}

func (m *Map) Update(k, v []byte) error {
	if m.perCPU {
		// Per-CPU maps need a buffer of value-size * num-CPUs.
		log.Panic("Per-CPU operations not implemented")
	}

	return UpdateMapEntry(m.fd, k, v)
}

func (m *Map) Get(k []byte) ([]byte, error) {
	if m.perCPU {
		// Per-CPU maps need a buffer of value-size * num-CPUs.
		log.Panic("Per-CPU operations not implemented")
	}

	return GetMapEntry(m.fd, k, m.ValueSize)
}

func (m *Map) Delete(k []byte) error {
	if m.perCPU {
		log.Panic("Per-CPU operations not implemented")
	}

	return DeleteMapEntry(m.fd, k, m.ValueSize)
}

type MapFD uint32

// INFO: BPF map 和程序作为内核资源只能通过文件描述符访问，其背后是内核中的匿名 inode，这带来了很多优点：
//  (1) 用户空间应用能够使用大部分文件描述符相关的 API
//  (2) 在 Unix socket 中传递文件描述符是透明的，等等
//  但同时，也有很多缺点：文件描述符受限于进程的生命周期，使得 map 共享之类的操作非常笨重。
//  为了解决这个问题，内核实现了一个最小内核空间 BPF 文件系统，BPF map 和 BPF 程序 都可以钉到（pin）这个文件系统内，这个过程称为 object pinning（钉住对象）
//  相应地，BPF 系统调用进行了扩展，添加了两个新命令，分别用于钉住（BPF_OBJ_PIN）一个对象和获取（BPF_OBJ_GET）一个被钉住的对象（pinned objects）
//  http://arthurchiao.art/blog/cilium-bpf-xdp-reference-guide-zh/#14-object-pinning%E9%92%89%E4%BD%8F%E5%AF%B9%E8%B1%A1
//  @see ObjGet(pathname)
type bpftoolMapMeta struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func RepinMap(name string, filename string) error {
	cmd := exec.Command("bpftool", "map", "list", "-j")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("bpftool map list failed: %v", err))
	}
	log.WithField("maps", string(out)).Debug("Got map metadata.")

	var mapMetas []bpftoolMapMeta
	err = json.Unmarshal(out, &mapMetas)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("bpftool returned bad JSON: %v", err))
	}

	for _, m := range mapMetas {
		if m.Name == name {
			// Found the map, try to repin it.
			cmd := exec.Command("bpftool", "map", "pin", "id", fmt.Sprint(m.ID), filename)
			_, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf(fmt.Sprintf("bpftool failed to repin map: %v", err))
			}
			return nil
		}
	}

	return os.ErrNotExist
}
