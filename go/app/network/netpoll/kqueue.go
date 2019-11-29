package netpoll

import "k8s-lx1036/app/network/internal"

// Poller ...
type Poller struct {
	fd            int
	asyncJobQueue internal.AsyncJobQueue
}
