package main

import (
	"github.com/cilium/ebpf"
	"golang.org/x/sys/unix"
)

// https://github.com/xdp-project/xdp-tutorial/blob/master/advanced03-AF_XDP/af_xdp_user.c

func main() {

	var xsksMap *ebpf.Map
	var xdpStatsMap *ebpf.Map

	// Allow unlimited locking of memory, so all memory needed for packet buffers can be locked.
	tmpLim := unix.Rlimit{
		Cur: unix.RLIM_INFINITY,
		Max: unix.RLIM_INFINITY,
	}
	err := unix.Setrlimit(unix.RLIMIT_MEMLOCK, &tmpLim)
	if err != nil {
	}

	/* Open and configure the AF_XDP (xsk) socket */

}
