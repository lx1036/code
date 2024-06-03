package policy

// PortProto is a pair of port number and protocol and is used as the
// value type in named port maps.
type PortProto struct {
    Port  uint16 // non-0
    Proto uint8  // 0 for any
}

// NamedPortMap maps port names to port numbers and protocols.
type NamedPortMap map[string]PortProto

// PortProtoSet is a set of unique PortProto values.
type PortProtoSet map[PortProto]struct{}

// Equal returns true if the PortProtoSets are equal.
func (pps PortProtoSet) Equal(other PortProtoSet) bool {
    if len(pps) != len(other) {
        return false
    }

    for port := range pps {
        if _, exists := other[port]; !exists {
            return false
        }
    }
    return true
}

// NamedPortMultiMap may have multiple entries for a name if multiple PODs
// define the same name with different values.
type NamedPortMultiMap map[string]PortProtoSet

// Equal returns true if the NamedPortMultiMaps are equal.
func (npm NamedPortMultiMap) Equal(other NamedPortMultiMap) bool {
    if len(npm) != len(other) {
        return false
    }

    for name, ports := range npm {
        if otherPorts, exists := other[name]; !exists || !ports.Equal(otherPorts) {
            return false
        }
    }

    return true
}
