package manager

import (
	"fmt"
	"net"
)

// defaultHealthProbeListener creates the default health probes listener bound to the given address
func defaultHealthProbeListener(addr string) (net.Listener, error) {
	if addr == "" || addr == "0" {
		return nil, nil
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("error listening on %s: %v", addr, err)
	}
	return ln, nil
}
