package lbmap

import (
	"io"
	"unsafe"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"
)

const (
	// Both inner maps are not being pinned into BPF fs.
	MaglevInner4MapName = "cilium_lb4_maglev_inner"
	MaglevInner6MapName = "cilium_lb6_maglev_inner"

	// Both outer maps are pinned though given we need to attach
	// inner maps into them.
	MaglevOuter4MapName = "cilium_lb4_maglev"
	MaglevOuter6MapName = "cilium_lb6_maglev"
)

var (
	MaglevOuter4Map     *bpf.Map
	MaglevOuter6Map     *bpf.Map
	maglevRecreatedIPv4 bool
	maglevRecreatedIPv6 bool
)

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type MaglevInnerKey struct{ Zero uint32 }

func (m *MaglevInnerKey) String() string {
	//TODO implement me
	panic("implement me")
}

func (m *MaglevInnerKey) GetKeyPtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (m *MaglevInnerKey) NewValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}

func (m *MaglevInnerKey) DeepCopyMapKey() bpf.MapKey {
	//TODO implement me
	panic("implement me")
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type MaglevInnerVal struct {
	BackendIDs []uint16
}

func (v *MaglevInnerVal) String() string {
	//TODO implement me
	panic("implement me")
}

func (v *MaglevInnerVal) GetValuePtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (v *MaglevInnerVal) DeepCopyMapValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}

func InitMaglevMaps(ipv4 bool, ipv6 bool) error {
	var err error

	dummyInnerMap := newInnerMaglevMap("cilium_lb_maglev_dummy")
	if err := dummyInnerMap.CreateUnpinned(); err != nil {
		return err
	}
	defer dummyInnerMap.Close()

	if ipv4 {
		if maglevRecreatedIPv4, err = deleteMapIfMNotMatch(MaglevOuter4MapName); err != nil {
			return err
		}
		MaglevOuter4Map = newOuterMaglevMap(MaglevOuter4MapName, dummyInnerMap)
		if _, err := MaglevOuter4Map.OpenOrCreate(); err != nil {
			return err
		}
	}

	return nil
}

func newInnerMaglevMap(name string) *bpf.Map {
	return bpf.NewMapWithOpts(
		name,
		bpf.MapTypeArray,
		&MaglevInnerKey{}, int(unsafe.Sizeof(MaglevInnerKey{})),
		&MaglevInnerVal{}, int(unsafe.Sizeof(uint16(0)))*option.Config.MaglevTableSize,
		1, 0, 0,
		bpf.ConvertKeyValue,
		&bpf.NewMapOpts{},
	)
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type MaglevOuterKey struct{ RevNatID uint16 }

func (k *MaglevOuterKey) String() string {
	//TODO implement me
	panic("implement me")
}

func (k *MaglevOuterKey) GetKeyPtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (k *MaglevOuterKey) NewValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}

func (k *MaglevOuterKey) DeepCopyMapKey() bpf.MapKey {
	//TODO implement me
	panic("implement me")
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type MaglevOuterVal struct{ FD uint32 }

func (v *MaglevOuterVal) String() string {
	//TODO implement me
	panic("implement me")
}

func (v *MaglevOuterVal) GetValuePtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (v *MaglevOuterVal) DeepCopyMapValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}

func newOuterMaglevMap(name string, innerMap *bpf.Map) *bpf.Map {
	return bpf.NewMap(
		name,
		bpf.MapTypeHashOfMaps,
		&MaglevOuterKey{}, int(unsafe.Sizeof(MaglevOuterKey{})),
		&MaglevOuterVal{}, int(unsafe.Sizeof(MaglevOuterVal{})),
		MaxEntries,
		0, uint32(innerMap.GetFd()),
		bpf.ConvertKeyValue,
	).WithPressureMetric()
}

// MaglevTableSize 可能变化，所以需要 check 要不要删除 cilium_lb4_maglev map
func deleteMapIfMNotMatch(mapName string) (bool, error) {
	deleteMap := false

	m, err := bpf.OpenMap(mapName)
	if err == nil {
		outerKey := &MaglevOuterKey{}
		if err := bpf.GetNextKey(m.GetFd(), nil, unsafe.Pointer(outerKey)); err == nil {
			outerVal, err := m.Lookup(outerKey)
			if err != nil {
				return false, err
			}
			v := outerVal.(*MaglevOuterVal)

			fd, err := bpf.MapFdFromID(int(v.FD))
			if err != nil {
				return false, err
			}
			info, err := bpf.GetMapInfoByFd(uint32(fd))
			if err != nil {
				return false, err
			}
			previousM := int(info.ValueSize) / int(unsafe.Sizeof(uint16(0)))
			if option.Config.MaglevTableSize != previousM {
				deleteMap = true
			}
		} else if err == io.EOF {
			// The map is empty. To be on the safe side, remove it to avoid M
			// mismatch. The removal is harmless, as no new entry can be
			// created while the initialization of the map has not returned.
			deleteMap = true
		} else {
			return false, err
		}
	}

	if deleteMap {
		log.WithField(logfields.BPFMapName, mapName).
			Info("Deleting Maglev outer map due to different M or empty map")
		if err := m.Unpin(); err != nil {
			return false, err
		}
	}

	return deleteMap, nil
}
