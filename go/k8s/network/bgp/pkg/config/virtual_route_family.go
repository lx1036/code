package config

// struct for container gobgp:state.
// Configured states of VRF.
type VrfState struct {
	// original -> gobgp:name
	// Unique name among all VRF instances.
	Name string `mapstructure:"name" json:"name,omitempty"`
	// original -> gobgp:id
	// Unique identifier among all VRF instances.
	Id uint32 `mapstructure:"id" json:"id,omitempty"`
	// original -> gobgp:rd
	// Route Distinguisher for this VRF.
	Rd string `mapstructure:"rd" json:"rd,omitempty"`
	// original -> gobgp:import-rt
	// List of import Route Targets for this VRF.
	ImportRtList []string `mapstructure:"import-rt-list" json:"import-rt-list,omitempty"`
	// original -> gobgp:export-rt
	// List of export Route Targets for this VRF.
	ExportRtList []string `mapstructure:"export-rt-list" json:"export-rt-list,omitempty"`
}

// struct for container gobgp:config.
// Configuration parameters for VRF.
type VrfConfig struct {
	// original -> gobgp:name
	// Unique name among all VRF instances.
	Name string `mapstructure:"name" json:"name,omitempty"`
	// original -> gobgp:id
	// Unique identifier among all VRF instances.
	Id uint32 `mapstructure:"id" json:"id,omitempty"`
	// original -> gobgp:rd
	// Route Distinguisher for this VRF.
	Rd string `mapstructure:"rd" json:"rd,omitempty"`
	// original -> gobgp:import-rt
	// List of import Route Targets for this VRF.
	ImportRtList []string `mapstructure:"import-rt-list" json:"import-rt-list,omitempty"`
	// original -> gobgp:export-rt
	// List of export Route Targets for this VRF.
	ExportRtList []string `mapstructure:"export-rt-list" json:"export-rt-list,omitempty"`
	// original -> gobgp:both-rt
	// List of both import and export Route Targets for this VRF. Each
	// configuration for import and export Route Targets will be preferred.
	BothRtList []string `mapstructure:"both-rt-list" json:"both-rt-list,omitempty"`
}

func (lhs *VrfConfig) Equal(rhs *VrfConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.Name != rhs.Name {
		return false
	}
	if lhs.Id != rhs.Id {
		return false
	}
	if lhs.Rd != rhs.Rd {
		return false
	}
	if len(lhs.ImportRtList) != len(rhs.ImportRtList) {
		return false
	}
	for idx, l := range lhs.ImportRtList {
		if l != rhs.ImportRtList[idx] {
			return false
		}
	}
	if len(lhs.ExportRtList) != len(rhs.ExportRtList) {
		return false
	}
	for idx, l := range lhs.ExportRtList {
		if l != rhs.ExportRtList[idx] {
			return false
		}
	}
	if len(lhs.BothRtList) != len(rhs.BothRtList) {
		return false
	}
	for idx, l := range lhs.BothRtList {
		if l != rhs.BothRtList[idx] {
			return false
		}
	}
	return true
}

// struct for container gobgp:vrf.
// VRF instance configurations on the local system.
type Vrf struct {
	// original -> gobgp:name
	// original -> gobgp:vrf-config
	// Configuration parameters for VRF.
	Config VrfConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> gobgp:vrf-state
	// Configured states of VRF.
	State VrfState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *Vrf) Equal(rhs *Vrf) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}
