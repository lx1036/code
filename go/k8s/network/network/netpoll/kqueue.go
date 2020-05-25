package netpoll

import "k8s-lx1036/k8s/network/network/internal"

// Poller ...
type Poller struct {
	fd            int
	asyncJobQueue internal.AsyncJobQueue
}
