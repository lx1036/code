//go:build linux

package main

import (
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
	"golang.org/x/sys/unix"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

// /root/linux-5.10.142/tools/testing/selftests/bpf/prog_tests/tcp_rtt.c

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -tags "linux" bpf tcprtt_sockops.c -- -I.

// go generate .
func main() {
	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGTERM)

	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal(err)
	}

	serverFd := startServer()

	// Find the path to a cgroup enabled to version 2
	cgroupPath, err := findCgroupPath()
	if err != nil {
		log.Fatal(err)
	}

	// Load pre-compiled programs and maps into the kernel.
	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		log.Fatalf("loading objects: %v", err)
	}
	defer objs.Close()

	// Attach ebpf program to a cgroupv2
	l, err := link.AttachCgroup(link.CgroupOptions{
		Path:    cgroupPath,
		Program: objs.bpfPrograms.BpfSockopsCb,
		Attach:  ebpf.AttachCGroupSockOps,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

}

func findCgroupPath() (string, error) {
	cgroupPath := "/sys/fs/cgroup"

	var st syscall.Statfs_t
	err := syscall.Statfs(cgroupPath, &st)
	if err != nil {
		return "", err
	}
	isCgroupV2Enabled := st.Type == unix.CGROUP2_SUPER_MAGIC
	if !isCgroupV2Enabled {
		cgroupPath = filepath.Join(cgroupPath, "unified")
	}
	return cgroupPath, nil
}

func startServer() int {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		log.Fatal(err)
	}

	setSocketTimeout(fd, 0)

	ip := net.ParseIP("127.0.0.1")
	sa := &unix.SockaddrInet4{
		Port: 8000,
		Addr: [4]byte{},
	}
	copy(sa.Addr[:], ip)
	err = unix.Bind(fd, sa)
	if err != nil {
		log.Fatal(err)
	}

	err = unix.Listen(fd, 1)
	if err != nil {
		log.Fatal(err)
	}

	return fd
}

func setSocketTimeout(fd, timeoutMs int) {
	var timeVal *unix.Timeval
	if timeoutMs > 0 {
		timeVal = &unix.Timeval{
			Sec:  int64(timeoutMs / 1000),
			Usec: int64(timeoutMs % 1000 * 1000),
		}
	} else {
		timeVal = &unix.Timeval{
			Sec: 3,
		}
	}

	err := unix.SetsockoptTimeval(fd, unix.SOL_SOCKET, unix.SO_RCVTIMEO, timeVal)
	if err != nil {
		log.Fatal(err)
	}

	err = unix.SetsockoptTimeval(fd, unix.SOL_SOCKET, unix.SO_SNDTIMEO, timeVal)
	if err != nil {
		log.Fatal(err)
	}
}
