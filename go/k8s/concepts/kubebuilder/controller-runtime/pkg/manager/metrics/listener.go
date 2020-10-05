package metrics

import (
	"fmt"
	"net"

	logf "sigs.k8s.io/controller-runtime/pkg/internal/log"
)

var log = logf.RuntimeLog.WithName("metrics")

// DefaultBindAddress sets the default bind address for the metrics listener
// The metrics is on by default.
var DefaultBindAddress = ":8080"

// NewListener creates a new TCP listener bound to the given address.
func NewListener(addr string) (net.Listener, error) {
	if addr == "" {
		// If the metrics bind address is empty, default to ":8080"
		addr = DefaultBindAddress
	}

	// Add a case to disable metrics altogether
	if addr == "0" {
		return nil, nil
	}

	log.Info("metrics server is starting to listen", "addr", addr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		er := fmt.Errorf("error listening on %s: %w", addr, err)
		log.Error(er, "metrics server failed to listen. You may want to disable the metrics server or use another port if it is due to conflicts")
		return nil, er
	}
	return ln, nil
}
