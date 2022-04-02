package v1

import (
	"bytes"
	"encoding/json"

	"github.com/golang/protobuf/jsonpb"
	gobgpapi "github.com/osrg/gobgp/v3/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BgpPeerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BgpPeer `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=bgpp,singular=bgppeer
// +kubebuilder:printcolumn:name="PeerAddress",type="string",JSONPath=".spec.peerAddress"
// +kubebuilder:printcolumn:name="PeerAsn",type="string",JSONPath=".spec.peerAsn"
// +kubebuilder:printcolumn:name="PeerPort",type="integer",JSONPath=".spec.peerPort"
// +kubebuilder:printcolumn:name="SourceAddress",type="string",JSONPath=".spec.sourceAddress"
// +kubebuilder:printcolumn:name="MyAsn",type="string",JSONPath=".spec.myAsn"
// +kubebuilder:printcolumn:name="SourcePort",type="integer",JSONPath=".spec.sourcePort"

type BgpPeer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BgpPeerSpec   `json:"spec,omitempty"`
	Status BgpPeerStatus `json:"status,omitempty"`
}

type BgpPeerSpec struct {
	// +kubebuilder:validation:Required
	PeerAddress string `json:"peerAddress,required"`

	// +kubebuilder:validation:Required
	PeerAsn int `json:"peerAsn,required"`

	PeerPort int `json:"peerPort,omitempty"`

	SourceAddress string `json:"sourceAddress,omitempty"`

	// +kubebuilder:validation:Required
	MyAsn int `json:"myAsn,required"`

	SourcePort int `json:"sourcePort,omitempty"`

	//Conf            *PeerConf        `json:"conf,omitempty"`
	//EbgpMultihop    *EbgpMultihop    `json:"ebgpMultihop,omitempty"`
	//Timers          *Timers          `json:"timers,omitempty"`
	//Transport       *Transport       `json:"transport,omitempty"`
	//GracefulRestart *GracefulRestart `json:"gracefulRestart,omitempty"`
	//AfiSafis        []*AfiSafi       `json:"afiSafis,omitempty"`

	NodeSelector *metav1.LabelSelector `json:"nodeSelector,omitempty"`
}

// ConvertToGoBgpPeer INFO: convert to pb message，可以借鉴！！！
func (spec BgpPeerSpec) ConvertToGoBgpPeer() (*gobgpapi.Peer, error) {
	spec.NodeSelector = nil

	jsonBytes, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}

	var result gobgpapi.Peer
	m := jsonpb.Unmarshaler{}
	err = m.Unmarshal(bytes.NewReader(jsonBytes), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

type BgpPeerStatus struct {
	NodesPeerStatus map[string]NodePeerStatus `json:"nodesPeerStatus,omitempty"`
}

type PeerConf struct {
	AuthPassword      string `json:"authPassword,omitempty"`
	Description       string `json:"description,omitempty"`
	LocalAs           uint32 `json:"localAs,omitempty"`
	NeighborAddress   string `json:"neighborAddress,omitempty"`
	PeerAs            uint32 `json:"peerAs,omitempty"`
	PeerGroup         string `json:"peerGroup,omitempty"`
	PeerType          uint32 `json:"peerType,omitempty"`
	RemovePrivateAs   string `json:"removePrivateAs,omitempty"`
	RouteFlapDamping  bool   `json:"routeFlapDamping,omitempty"`
	SendCommunity     uint32 `json:"sendCommunity,omitempty"`
	NeighborInterface string `json:"neighborInterface,omitempty"`
	Vrf               string `json:"vrf,omitempty"`
	AllowOwnAs        uint32 `json:"allowOwnAs,omitempty"`
	ReplacePeerAs     bool   `json:"replacePeerAs,omitempty"`
	AdminDown         bool   `json:"adminDown,omitempty"`
}

type EbgpMultihop struct {
	Enabled     bool   `json:"enabled,omitempty"`
	MultihopTtl uint32 `json:"multihopTtl,omitempty"`
}

type Timers struct {
	Config *TimersConfig `json:"config,omitempty"`
}

// https://stackoverflow.com/questions/21151765/cannot-unmarshal-string-into-go-value-of-type-int64
type TimersConfig struct {
	ConnectRetry                 string `json:"connectRetry,omitempty"`
	HoldTime                     string `json:"holdTime,omitempty"`
	KeepaliveInterval            string `json:"keepaliveInterval,omitempty"`
	MinimumAdvertisementInterval string `json:"minimumAdvertisementInterval,omitempty"`
}

type Transport struct {
	MtuDiscovery  bool   `json:"mtuDiscovery,omitempty"`
	PassiveMode   bool   `json:"passiveMode,omitempty"`
	RemoteAddress string `json:"remoteAddress,omitempty"`
	RemotePort    uint32 `json:"remotePort,omitempty"`
	TcpMss        uint32 `json:"tcpMss,omitempty"`
}

type AfiSafi struct {
	MpGracefulRestart *MpGracefulRestart `json:"mpGracefulRestart,omitempty"`
	Config            *AfiSafiConfig     `json:"config,omitempty"`
	AddPaths          *AddPaths          `json:"addPaths,omitempty"`
}

type MpGracefulRestartConfig struct {
	Enabled bool `json:"enabled,omitempty"`
}

type MpGracefulRestart struct {
	Config *MpGracefulRestartConfig `json:"config,omitempty"`
}

type Family struct {
	Afi  string `json:"afi,omitempty"`
	Safi string `json:"safi,omitempty"`
}

type AfiSafiConfig struct {
	Family  *Family `json:"family,omitempty"`
	Enabled bool    `json:"enabled,omitempty"`
}

type AddPathsConfig struct {
	Receive bool   `json:"receive,omitempty"`
	SendMax uint32 `json:"sendMax,omitempty"`
}

type AddPaths struct {
	Config *AddPathsConfig `json:"config,omitempty"`
}

type Message struct {
	Notification   string `json:"notification,omitempty"`
	Update         string `json:"update,omitempty"`
	Open           string `json:"open,omitempty"`
	Keepalive      string `json:"keepalive,omitempty"`
	Refresh        string `json:"refresh,omitempty"`
	Discarded      string `json:"discarded,omitempty"`
	Total          string `json:"total,omitempty"`
	WithdrawUpdate string `json:"withdrawUpdate,omitempty"`
	WithdrawPrefix string `json:"withdrawPrefix,omitempty"`
}

type Messages struct {
	Received *Message `json:"received,omitempty"`
	Sent     *Message `json:"sent,omitempty"`
}

type Queues struct {
	Input  uint32 `json:"input,omitempty"`
	Output uint32 `json:"output,omitempty"`
}

type PeerState struct {
	AuthPassword     string    `json:"authPassword,omitempty"`
	Description      string    `json:"description,omitempty"`
	LocalAs          uint32    `json:"localAs,omitempty"`
	Messages         *Messages `json:"messages,omitempty"`
	NeighborAddress  string    `json:"neighborAddress,omitempty"`
	PeerAs           uint32    `json:"peerAs,omitempty"`
	PeerGroup        string    `json:"peerGroup,omitempty"`
	PeerType         uint32    `json:"peerType,omitempty"`
	Queues           *Queues   `json:"queues,omitempty"`
	RemovePrivateAs  uint32    `json:"removePrivateAs,omitempty"`
	RouteFlapDamping bool      `json:"routeFlapDamping,omitempty"`
	SendCommunity    uint32    `json:"sendCommunity,omitempty"`
	SessionState     string    `json:"sessionState,omitempty"`
	AdminState       string    `json:"adminState,omitempty"`
	OutQ             uint32    `json:"outQ,omitempty"`
	Flops            uint32    `json:"flops,omitempty"`
	RouterId         string    `json:"routerId,omitempty"`
}

type TimersState struct {
	ConnectRetry                 string `json:"connectRetry,omitempty"`
	HoldTime                     string `json:"holdTime,omitempty"`
	KeepaliveInterval            string `json:"keepaliveInterval,omitempty"`
	MinimumAdvertisementInterval string `json:"minimumAdvertisementInterval,omitempty"`
	NegotiatedHoldTime           string `json:"negotiatedHoldTime,omitempty"`
	Uptime                       string `json:"uptime,omitempty"`
	Downtime                     string `json:"downtime,omitempty"`
}

type NodePeerStatus struct {
	PeerState   PeerState   `json:"peerState,omitempty"`
	TimersState TimersState `json:"timersState,omitempty"`
}
