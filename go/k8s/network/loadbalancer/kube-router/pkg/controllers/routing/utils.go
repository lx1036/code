package routing

import (
	"fmt"
	"net"
	"strings"

	"github.com/golang/protobuf/ptypes"
	gobgpapi "github.com/osrg/gobgp/v3/api"
	"github.com/vishvananda/netlink"
)

// parseBGPPath takes in a GoBGP Path and parses out the destination subnet and the next hop from its attributes.
// If successful, it will return the destination of the BGP path as a subnet form and the next hop. If it
// can't parse the destination or the next hop IP, it returns an error.
func parseBGPPath(path *gobgpapi.Path) (*net.IPNet, net.IP, error) {
	nextHop, err := parseBGPNextHop(path)
	if err != nil {
		return nil, nil, err
	}

	nlri := path.GetNlri()
	var prefix gobgpapi.IPAddressPrefix
	err = ptypes.UnmarshalAny(nlri, &prefix)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid nlri in advertised path")
	}
	dstSubnet, err := netlink.ParseIPNet(prefix.Prefix + "/" + fmt.Sprint(prefix.PrefixLen))
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't parse IP subnet from nlri advertised path")
	}

	return dstSubnet, nextHop, nil
}

// parseBGPNextHop takes in a GoBGP Path and parses out the destination's next hop from its attributes. If it
// can't parse a next hop IP from the GoBGP Path, it returns an error.
func parseBGPNextHop(path *gobgpapi.Path) (net.IP, error) {
	for _, pAttr := range path.GetPattrs() {
		var value ptypes.DynamicAny
		if err := ptypes.UnmarshalAny(pAttr, &value); err != nil {
			return nil, fmt.Errorf("failed to unmarshal path attribute: %s", err)
		}
		// nolint:gocritic // We can't change this to an if condition because it is a .(type) expression
		switch a := value.Message.(type) {
		case *gobgpapi.NextHopAttribute:
			nextHop := net.ParseIP(a.NextHop).To4()
			if nextHop == nil {
				if nextHop = net.ParseIP(a.NextHop).To16(); nextHop == nil {
					return nil, fmt.Errorf("invalid nextHop address: %s", a.NextHop)
				}
			}
			return nextHop, nil
		}
	}

	return nil, fmt.Errorf("could not parse next hop received from GoBGP for path: %s", path)
}

// generateTunnelName will generate a name for a tunnel interface given a node IP
// for example, if the node IP is 10.0.0.1 the tunnel interface will be named tun-10001
// Since linux restricts interface names to 15 characters, if length of a node IP
// is greater than 12 (after removing "."), then the interface name is tunXYZ
// as opposed to tun-XYZ
func generateTunnelName(nodeIP string) string {
	hash := strings.ReplaceAll(nodeIP, ".", "")

	// nolint:gomnd // this number becomes less obvious when made a constant
	if len(hash) < 12 {
		return "tun-" + hash
	}

	return "tun" + hash
}
