package subnet

import (
	"encoding/json"
	"time"

	"k8s-lx1036/k8s/network/cni/flannel/pkg/ip"
)

type LeaseAttrs struct {
	PublicIP      ip.IP4
	BackendType   string          `json:",omitempty"`
	BackendData   json.RawMessage `json:",omitempty"`
	BackendV6Data json.RawMessage `json:",omitempty"`
}

type Lease struct {
	EnableIPv4 bool
	Subnet     ip.IP4Net
	Attrs      LeaseAttrs
	Expiration time.Time

	Asof int64
}

func (l *Lease) Key() string {
	return MakeSubnetKey(l.Subnet)
}

func MakeSubnetKey(sn ip.IP4Net) string {
	return sn.StringSep(".", "-")
}

type LeaseWatchResult struct {
	// Either Events or Snapshot will be set.  If Events is empty, it means
	// the cursor was out of range and Snapshot contains the current list
	// of items, even if empty.
	Events   []Event     `json:"events"`
	Snapshot []Lease     `json:"snapshot"`
	Cursor   interface{} `json:"cursor"`
}
