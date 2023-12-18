//go:build linux

package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/perf"
	"github.com/cilium/ebpf/rlimit"
	"github.com/coreos/go-iptables/iptables"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -tags "linux" -type tcp_notifier bpf tcpnotify.c -- -I.

// /root/linux-5.10.142/tools/testing/selftests/bpf/test_tcpnotify_user.c

const (
	TESTPORT = 12877
)

var (
	rxCallbacks = 0
)

// go run .
func main() {
	logrus.SetReportCaller(true)

	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		logrus.Fatal(err)
	}

	// Find the path to a cgroup enabled to version 2
	cgroupPath, err := findCgroupPath()
	if err != nil {
		logrus.Fatal(err)
	}

	// Load pre-compiled programs and maps into the kernel.
	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		logrus.Fatalf("loading objects: %v", err)
	}
	defer objs.Close()

	// Attach ebpf program to a cgroupv2
	l, err := link.AttachCgroup(link.CgroupOptions{
		Path:    cgroupPath,
		Program: objs.bpfPrograms.BpfSockopsCb,
		Attach:  ebpf.AttachCGroupSockOps,
	})
	if err != nil {
		logrus.Fatal(err)
	}
	defer l.Close()

	go readFromPerfEventMap(objs)

	ipt, err := iptables.New()
	if err != nil {
		logrus.Fatal(err)
	}

	dropRule := []string{"-p", "tcp", "--dport", strconv.Itoa(TESTPORT), "-j", "DROP"}
	err = ipt.AppendUnique("filter", "INPUT", dropRule...)
	if err != nil {
		logrus.Fatal(err)
	}

	output, err := exec.Command("nc", "127.0.0.1", strconv.Itoa(TESTPORT), "-v").Output()
	if err != nil {
		logrus.Fatal(err)
	} else {
		logrus.Infof("exec output: %s", string(output))
	}

	time.Sleep(time.Second * 5)

	err = ipt.DeleteIfExists("filter", "INPUT", dropRule...)
	if err != nil {
		logrus.Fatal(err)
	}

	key := uint32(0)
	var result bpfTcpnotifyGlobals
	err = objs.bpfMaps.GlobalMap.Lookup(key, result)
	if err != nil {
		logrus.Fatal(err)
	}

	time.Sleep(time.Second * 2)

	logrus.Infof("bpfTcpnotifyGlobals: %+v, rxCallbacks: %d", result, rxCallbacks)
}

func readFromPerfEventMap(objs bpfObjects) {
	// Open a perf event reader from userspace on the PERF_EVENT_ARRAY map
	// described in the eBPF C program.
	rd, err := perf.NewReader(objs.bpfMaps.PerfEventMap, os.Getpagesize())
	if err != nil {
		logrus.Fatalf("creating perf event reader: %s", err)
	}
	defer rd.Close()

	// bpfEvent is generated by bpf2go.
	var event bpfTcpNotifier
	for {
		record, err := rd.Read()
		if err != nil {
			if errors.Is(err, perf.ErrClosed) {
				return
			}
			logrus.Printf("reading from perf event reader: %s", err)
			continue
		}

		if record.LostSamples != 0 {
			logrus.Printf("perf event ring buffer full, dropped %d samples", record.LostSamples)
			continue
		}

		// Parse the perf event entry into a bpfEvent structure.
		if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &event); err != nil {
			logrus.Printf("parsing perf event: %s", err)
			continue
		}

		if event.Type != uint8(0xde) || event.Subtype != uint8(0xad) || event.Source != uint8(0xbe) || event.Hash != uint8(0xef) {
			continue
		}

		rxCallbacks++
	}
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
