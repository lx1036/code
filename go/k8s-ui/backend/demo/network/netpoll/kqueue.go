package netpoll

import "k8s-lx1036/k8s-ui/backend/demo/network/internal"

// Poller ...
type Poller struct {
	fd            int
	asyncJobQueue internal.AsyncJobQueue
}
