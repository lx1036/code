// Code generated by "stringer -type=BGPAttrType"; DO NOT EDIT.

package bgp

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[BGP_ATTR_TYPE_ORIGIN-1]
	_ = x[BGP_ATTR_TYPE_AS_PATH-2]
	_ = x[BGP_ATTR_TYPE_NEXT_HOP-3]
	_ = x[BGP_ATTR_TYPE_MULTI_EXIT_DISC-4]
	_ = x[BGP_ATTR_TYPE_LOCAL_PREF-5]
	_ = x[BGP_ATTR_TYPE_ATOMIC_AGGREGATE-6]
	_ = x[BGP_ATTR_TYPE_AGGREGATOR-7]
	_ = x[BGP_ATTR_TYPE_COMMUNITIES-8]
	_ = x[BGP_ATTR_TYPE_ORIGINATOR_ID-9]
	_ = x[BGP_ATTR_TYPE_CLUSTER_LIST-10]
	_ = x[BGP_ATTR_TYPE_MP_REACH_NLRI-14]
	_ = x[BGP_ATTR_TYPE_MP_UNREACH_NLRI-15]
	_ = x[BGP_ATTR_TYPE_EXTENDED_COMMUNITIES-16]
	_ = x[BGP_ATTR_TYPE_AS4_PATH-17]
	_ = x[BGP_ATTR_TYPE_AS4_AGGREGATOR-18]
	_ = x[BGP_ATTR_TYPE_PMSI_TUNNEL-22]
	_ = x[BGP_ATTR_TYPE_TUNNEL_ENCAP-23]
	_ = x[BGP_ATTR_TYPE_IP6_EXTENDED_COMMUNITIES-25]
	_ = x[BGP_ATTR_TYPE_AIGP-26]
	_ = x[BGP_ATTR_TYPE_LS-29]
	_ = x[BGP_ATTR_TYPE_LARGE_COMMUNITY-32]
}

const (
	_BGPAttrType_name_0 = "BGP_ATTR_TYPE_ORIGINBGP_ATTR_TYPE_AS_PATHBGP_ATTR_TYPE_NEXT_HOPBGP_ATTR_TYPE_MULTI_EXIT_DISCBGP_ATTR_TYPE_LOCAL_PREFBGP_ATTR_TYPE_ATOMIC_AGGREGATEBGP_ATTR_TYPE_AGGREGATORBGP_ATTR_TYPE_COMMUNITIESBGP_ATTR_TYPE_ORIGINATOR_IDBGP_ATTR_TYPE_CLUSTER_LIST"
	_BGPAttrType_name_1 = "BGP_ATTR_TYPE_MP_REACH_NLRIBGP_ATTR_TYPE_MP_UNREACH_NLRIBGP_ATTR_TYPE_EXTENDED_COMMUNITIESBGP_ATTR_TYPE_AS4_PATHBGP_ATTR_TYPE_AS4_AGGREGATOR"
	_BGPAttrType_name_2 = "BGP_ATTR_TYPE_PMSI_TUNNELBGP_ATTR_TYPE_TUNNEL_ENCAP"
	_BGPAttrType_name_3 = "BGP_ATTR_TYPE_IP6_EXTENDED_COMMUNITIESBGP_ATTR_TYPE_AIGP"
	_BGPAttrType_name_4 = "BGP_ATTR_TYPE_LS"
	_BGPAttrType_name_5 = "BGP_ATTR_TYPE_LARGE_COMMUNITY"
)

var (
	_BGPAttrType_index_0 = [...]uint8{0, 20, 41, 63, 92, 116, 146, 170, 195, 222, 248}
	_BGPAttrType_index_1 = [...]uint8{0, 27, 56, 90, 112, 140}
	_BGPAttrType_index_2 = [...]uint8{0, 25, 51}
	_BGPAttrType_index_3 = [...]uint8{0, 38, 56}
)

func (i BGPAttrType) String() string {
	switch {
	case 1 <= i && i <= 10:
		i -= 1
		return _BGPAttrType_name_0[_BGPAttrType_index_0[i]:_BGPAttrType_index_0[i+1]]
	case 14 <= i && i <= 18:
		i -= 14
		return _BGPAttrType_name_1[_BGPAttrType_index_1[i]:_BGPAttrType_index_1[i+1]]
	case 22 <= i && i <= 23:
		i -= 22
		return _BGPAttrType_name_2[_BGPAttrType_index_2[i]:_BGPAttrType_index_2[i+1]]
	case 25 <= i && i <= 26:
		i -= 25
		return _BGPAttrType_name_3[_BGPAttrType_index_3[i]:_BGPAttrType_index_3[i+1]]
	case i == 29:
		return _BGPAttrType_name_4
	case i == 32:
		return _BGPAttrType_name_5
	default:
		return "BGPAttrType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
