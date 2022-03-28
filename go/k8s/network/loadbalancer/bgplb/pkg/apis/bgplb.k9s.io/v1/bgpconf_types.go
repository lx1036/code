package v1

import (
	"bytes"
	"encoding/json"

	"github.com/golang/protobuf/jsonpb"
	gobgpapi "github.com/osrg/gobgp/v3/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// INFO: @see https://github.com/kubesphere/openelb/blob/master/doc/zh/bgp_config.md

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BgpConfList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BgpConf `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=bgpconfs,scope=Cluster,shortName=conf,singular=bgpconf
// +kubebuilder:printcolumn:name="As",type="integer",JSONPath=".spec.as"
// +kubebuilder:printcolumn:name="ListenPort",type="integer",JSONPath=".spec.listenPort"
// +kubebuilder:printcolumn:name="RouterId",type="string",JSONPath=".spec.routerId"

type BgpConf struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BgpConfSpec   `json:"spec,omitempty"`
	Status BgpConfStatus `json:"status,omitempty"`
}

// Configuration parameters relating to the global BGP router.
type BgpConfSpec struct {
	As               uint32            `json:"as,omitempty"`
	AsPerRack        map[string]uint32 `json:"asPerRack,omitempty"`
	RouterId         string            `json:"routerId,omitempty"`
	ListenPort       int32             `json:"listenPort,omitempty"`
	ListenAddresses  []string          `json:"listenAddresses,omitempty"`
	Families         []uint32          `json:"families,omitempty"`
	UseMultiplePaths bool              `json:"useMultiplePaths,omitempty"`
	GracefulRestart  *GracefulRestart  `json:"gracefulRestart,omitempty"`
}

// ConvertToGoBGPGlobalConf INFO: convert to pb message，可以借鉴！！！
func (spec BgpConfSpec) ConvertToGoBGPGlobalConf() (*gobgpapi.Global, error) {
	spec.AsPerRack = nil

	jsonBytes, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}

	var result gobgpapi.Global
	m := &jsonpb.Unmarshaler{}
	err = m.Unmarshal(bytes.NewReader(jsonBytes), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

type GracefulRestart struct {
	Enabled             bool   `json:"enabled,omitempty"`
	RestartTime         uint32 `json:"restartTime,omitempty"`
	HelperOnly          bool   `json:"helperOnly,omitempty"`
	DeferralTime        uint32 `json:"deferralTime,omitempty"`
	NotificationEnabled bool   `json:"notificationEnabled,omitempty"`
	LonglivedEnabled    bool   `json:"longlivedEnabled,omitempty"`
	StaleRoutesTime     uint32 `json:"staleRoutesTime,omitempty"`
	PeerRestartTime     uint32 `json:"peerRestartTime,omitempty"`
	PeerRestarting      bool   `json:"peerRestarting,omitempty"`
	LocalRestarting     bool   `json:"localRestarting,omitempty"`
	Mode                string `json:"mode,omitempty"`
}

type BgpConfStatus struct {
	NodesConfStatus map[string]NodeConfStatus `json:"nodesConfStatus,omitempty"`
}

type NodeConfStatus struct {
	RouterId string `json:"routerId,omitempty"`
	As       uint32 `json:"as,omitempty"`
}
