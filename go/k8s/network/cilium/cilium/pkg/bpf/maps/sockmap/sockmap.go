package sockmap

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"unsafe"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf"
)

const (
	mapName = "cilium_sock_ops"

	// MaxEntries represents the maximum number of endpoints in the map
	MaxEntries = 65535
)

var (
	buildMap sync.Once
	// SockMap represents the BPF map for sockets
	SockMap *bpf.Map
)

func SockmapCreate() {
	if err := CreateWithName(mapName); err != nil {
		log.WithError(err).Warning("Unable to open or create socket map")
	}
}

func CreateWithName(name string) error {
	buildMap.Do(func() {
		SockMap = bpf.NewMap(name,
			bpf.MapTypeSockHash,
			&SockmapKey{},
			int(unsafe.Sizeof(SockmapKey{})),
			&SockmapValue{},
			4,
			MaxEntries,
			0, 0,
			bpf.ConvertKeyValue,
		)
	})

	_, err := SockMap.OpenOrCreate()
	return err
}

// SockmapKey is the 5-tuple used to lookup a socket
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type SockmapKey struct {
	//DIP    types.IPv6 `align:"$union0"`
	//SIP    types.IPv6 `align:"$union1"`
	Family uint8  `align:"family"`
	Pad7   uint8  `align:"pad7"`
	Pad8   uint16 `align:"pad8"`
	SPort  uint32 `align:"sport"`
	DPort  uint32 `align:"dport"`
}

// SockmapValue is the fd of a socket
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type SockmapValue struct {
	fd uint32
}
