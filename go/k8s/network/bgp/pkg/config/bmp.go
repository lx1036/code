package config

import "fmt"

// INFO: BMP: BGP Monitoring Protocol, provides a convenient interface for obtaining route views
//  https://github.com/osrg/gobgp/blob/master/docs/sources/bmp.md

// typedef for identity gobgp:bmp-route-monitoring-policy-type.
type BmpRouteMonitoringPolicyType string

const (
	BMP_ROUTE_MONITORING_POLICY_TYPE_PRE_POLICY  BmpRouteMonitoringPolicyType = "pre-policy"
	BMP_ROUTE_MONITORING_POLICY_TYPE_POST_POLICY BmpRouteMonitoringPolicyType = "post-policy"
	BMP_ROUTE_MONITORING_POLICY_TYPE_BOTH        BmpRouteMonitoringPolicyType = "both"
	BMP_ROUTE_MONITORING_POLICY_TYPE_LOCAL_RIB   BmpRouteMonitoringPolicyType = "local-rib"
	BMP_ROUTE_MONITORING_POLICY_TYPE_ALL         BmpRouteMonitoringPolicyType = "all"
)

var BmpRouteMonitoringPolicyTypeToIntMap = map[BmpRouteMonitoringPolicyType]int{
	BMP_ROUTE_MONITORING_POLICY_TYPE_PRE_POLICY:  0,
	BMP_ROUTE_MONITORING_POLICY_TYPE_POST_POLICY: 1,
	BMP_ROUTE_MONITORING_POLICY_TYPE_BOTH:        2,
	BMP_ROUTE_MONITORING_POLICY_TYPE_LOCAL_RIB:   3,
	BMP_ROUTE_MONITORING_POLICY_TYPE_ALL:         4,
}

var IntToBmpRouteMonitoringPolicyTypeMap = map[int]BmpRouteMonitoringPolicyType{
	0: BMP_ROUTE_MONITORING_POLICY_TYPE_PRE_POLICY,
	1: BMP_ROUTE_MONITORING_POLICY_TYPE_POST_POLICY,
	2: BMP_ROUTE_MONITORING_POLICY_TYPE_BOTH,
	3: BMP_ROUTE_MONITORING_POLICY_TYPE_LOCAL_RIB,
	4: BMP_ROUTE_MONITORING_POLICY_TYPE_ALL,
}

func (v BmpRouteMonitoringPolicyType) Validate() error {
	if _, ok := BmpRouteMonitoringPolicyTypeToIntMap[v]; !ok {
		return fmt.Errorf("invalid BmpRouteMonitoringPolicyType: %s", v)
	}
	return nil
}

func (v BmpRouteMonitoringPolicyType) ToInt() int {
	i, ok := BmpRouteMonitoringPolicyTypeToIntMap[v]
	if !ok {
		return -1
	}
	return i
}

// struct for container gobgp:state.
// Configuration parameters relating to BMP server.
type BmpServerState struct {
	// original -> gobgp:address
	// gobgp:address's original type is inet:ip-address.
	// Reference to the address of the BMP server used as
	// a key in the BMP server list.
	Address string `mapstructure:"address" json:"address,omitempty"`
	// original -> gobgp:port
	// Reference to the port of the BMP server.
	Port uint32 `mapstructure:"port" json:"port,omitempty"`
	// original -> gobgp:route-monitoring-policy
	RouteMonitoringPolicy BmpRouteMonitoringPolicyType `mapstructure:"route-monitoring-policy" json:"route-monitoring-policy,omitempty"`
	// original -> gobgp:statistics-timeout
	// Interval seconds of statistics messages sent to BMP server.
	StatisticsTimeout uint16 `mapstructure:"statistics-timeout" json:"statistics-timeout,omitempty"`
	// original -> gobgp:route-mirroring-enabled
	// gobgp:route-mirroring-enabled's original type is boolean.
	// Enable feature for mirroring of received BGP messages
	// mainly for debugging purpose.
	RouteMirroringEnabled bool `mapstructure:"route-mirroring-enabled" json:"route-mirroring-enabled,omitempty"`
	// original -> gobgp:sys-name
	// Reference to the SysName of the BMP server.
	SysName string `mapstructure:"sys-name" json:"sys-name,omitempty"`
	// original -> gobgp:sys-descr
	// Reference to the SysDescr of the BMP server.
	SysDescr string `mapstructure:"sys-descr" json:"sys-descr,omitempty"`
}

// struct for container gobgp:config.
// Configuration parameters relating to BMP server.
type BmpServerConfig struct {
	// original -> gobgp:address
	// gobgp:address's original type is inet:ip-address.
	// Reference to the address of the BMP server used as
	// a key in the BMP server list.
	Address string `mapstructure:"address" json:"address,omitempty"`
	// original -> gobgp:port
	// Reference to the port of the BMP server.
	Port uint32 `mapstructure:"port" json:"port,omitempty"`
	// original -> gobgp:route-monitoring-policy
	RouteMonitoringPolicy BmpRouteMonitoringPolicyType `mapstructure:"route-monitoring-policy" json:"route-monitoring-policy,omitempty"`
	// original -> gobgp:statistics-timeout
	// Interval seconds of statistics messages sent to BMP server.
	StatisticsTimeout uint16 `mapstructure:"statistics-timeout" json:"statistics-timeout,omitempty"`
	// original -> gobgp:route-mirroring-enabled
	// gobgp:route-mirroring-enabled's original type is boolean.
	// Enable feature for mirroring of received BGP messages
	// mainly for debugging purpose.
	RouteMirroringEnabled bool `mapstructure:"route-mirroring-enabled" json:"route-mirroring-enabled,omitempty"`
	// original -> gobgp:sys-name
	// Reference to the SysName of the BMP server.
	SysName string `mapstructure:"sys-name" json:"sys-name,omitempty"`
	// original -> gobgp:sys-descr
	// Reference to the SysDescr of the BMP server.
	SysDescr string `mapstructure:"sys-descr" json:"sys-descr,omitempty"`
}

func (lhs *BmpServerConfig) Equal(rhs *BmpServerConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.Address != rhs.Address {
		return false
	}
	if lhs.Port != rhs.Port {
		return false
	}
	if lhs.RouteMonitoringPolicy != rhs.RouteMonitoringPolicy {
		return false
	}
	if lhs.StatisticsTimeout != rhs.StatisticsTimeout {
		return false
	}
	if lhs.RouteMirroringEnabled != rhs.RouteMirroringEnabled {
		return false
	}
	if lhs.SysName != rhs.SysName {
		return false
	}
	if lhs.SysDescr != rhs.SysDescr {
		return false
	}
	return true
}

// struct for container gobgp:bmp-server.
// List of BMP servers configured on the local system.
type BmpServer struct {
	// original -> gobgp:address
	// original -> gobgp:bmp-server-config
	// Configuration parameters relating to BMP server.
	Config BmpServerConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> gobgp:bmp-server-state
	// Configuration parameters relating to BMP server.
	State BmpServerState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *BmpServer) Equal(rhs *BmpServer) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}
