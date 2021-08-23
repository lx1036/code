package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/lni/dragonboat/v3"
	"github.com/lni/dragonboat/v3/config"
	"github.com/lni/dragonboat/v3/logger"
	"github.com/lni/goutils/syncutil"
)

var (
	// initial nodes count is fixed to three, their addresses are also fixed
	// these are the initial member nodes of the Raft cluster.
	addresses = []string{
		"localhost:63001",
		"localhost:63002",
		"localhost:63003",
	}
	errNotMembershipChange = errors.New("not a membership change request")
)

const (
	exampleClusterID uint64 = 128
)

// go run . --nodeid=1
// go run . --nodeid=2
// go run . --nodeid=3
func main() {
	nodeID := flag.Int("nodeid", 1, "NodeID to use")
	addr := flag.String("addr", "", "Nodehost address")
	join := flag.Bool("join", false, "Joining a new node")
	flag.Parse()
	if len(*addr) == 0 && *nodeID != 1 && *nodeID != 2 && *nodeID != 3 {
		fmt.Fprintf(os.Stderr, "node id must be 1, 2 or 3 when address is not specified\n")
		os.Exit(1)
	}
	// https://github.com/golang/go/issues/17393
	if runtime.GOOS == "darwin" {
		signal.Ignore(syscall.Signal(0xd))
	}
	initialMembers := make(map[uint64]string)
	// when joining a new node which is not an initial members, the initialMembers
	// map should be empty.
	// when restarting a node that is not a member of the initial nodes, you can
	// leave the initialMembers to be empty. we still populate the initialMembers
	// here for simplicity.
	if !*join {
		for idx, v := range addresses {
			// key is the NodeID, NodeID is not allowed to be 0
			// value is the raft address
			initialMembers[uint64(idx+1)] = v
		}
	}
	var nodeAddr string
	// for simplicity, in this example program, addresses of all those 3 initial
	// raft members are hard coded. when address is not specified on the command
	// line, we assume the node being launched is an initial raft member.
	if len(*addr) != 0 {
		nodeAddr = *addr
	} else {
		nodeAddr = initialMembers[uint64(*nodeID)]
	}
	fmt.Fprintf(os.Stdout, "node address: %s\n", nodeAddr)
	// change the log verbosity
	logger.GetLogger("raft").SetLevel(logger.ERROR)
	logger.GetLogger("rsm").SetLevel(logger.WARNING)
	logger.GetLogger("transport").SetLevel(logger.WARNING)
	logger.GetLogger("grpc").SetLevel(logger.WARNING)
	// config for raft node
	// See GoDoc for all available options
	rc := config.Config{
		// ClusterID and NodeID of the raft node
		NodeID:    uint64(*nodeID),
		ClusterID: exampleClusterID,
		// In this example, we assume the end-to-end round trip time (RTT) between
		// NodeHost instances (on different machines, VMs or containers) are 200
		// millisecond, it is set in the RTTMillisecond field of the
		// config.NodeHostConfig instance below.
		// ElectionRTT is set to 10 in this example, it determines that the node
		// should start an election if there is no heartbeat from the leader for
		// 10 * RTT time intervals.
		ElectionRTT: 10,
		// HeartbeatRTT is set to 1 in this example, it determines that when the
		// node is a leader, it should broadcast heartbeat messages to its followers
		// every such 1 * RTT time interval.
		HeartbeatRTT: 1,
		CheckQuorum:  true,
		// SnapshotEntries determines how often should we take a snapshot of the
		// replicated state machine, it is set to 10 her which means a snapshot
		// will be captured for every 10 applied proposals (writes).
		// In your real world application, it should be set to much higher values
		// You need to determine a suitable value based on how much space you are
		// willing use on Raft Logs, how fast can you capture a snapshot of your
		// replicated state machine, how often such snapshot is going to be used
		// etc.
		SnapshotEntries: 10,
		// Once a snapshot is captured and saved, how many Raft entries already
		// covered by the new snapshot should be kept. This is useful when some
		// followers are just a little bit left behind, with such overhead Raft
		// entries, the leaders can send them regular entries rather than the full
		// snapshot image.
		CompactionOverhead: 5,
	}
	datadir := filepath.Join(
		"raft",
		"helloworld",
		fmt.Sprintf("node%d", *nodeID))
	// config for the nodehost
	// See GoDoc for all available options
	// by default, insecure transport is used, you can choose to use Mutual TLS
	// Authentication to authenticate both servers and clients. To use Mutual
	// TLS Authentication, set the MutualTLS field in NodeHostConfig to true, set
	// the CAFile, CertFile and KeyFile fields to point to the path of your CA
	// file, certificate and key files.
	nhc := config.NodeHostConfig{
		// WALDir is the directory to store the WAL of all Raft Logs. It is
		// recommended to use Enterprise SSDs with good fsync() performance
		// to get the best performance. A few SSDs we tested or known to work very
		// well
		// Recommended SATA SSDs -
		// Intel S3700, Intel S3710, Micron 500DC
		// Other SATA enterprise class SSDs with power loss protection
		// Recommended NVME SSDs -
		// Most enterprise NVME currently available on the market.
		// SSD to avoid -
		// Consumer class SSDs, no matter whether they are SATA or NVME based, as
		// they usually have very poor fsync() performance.
		//
		// You can use the pg_test_fsync tool shipped with PostgreSQL to test the
		// fsync performance of your WAL disk. It is recommended to use SSDs with
		// fsync latency of well below 1 millisecond.
		//
		// Note that this is only for storing the WAL of Raft Logs, it is size is
		// usually pretty small, 64GB per NodeHost is usually more than enough.
		//
		// If you just have one disk in your system, just set WALDir and NodeHostDir
		// to the same location.
		WALDir: datadir,
		// NodeHostDir is where everything else is stored.
		NodeHostDir: datadir,
		// RTTMillisecond is the average round trip time between NodeHosts (usually
		// on two machines/vms), it is in millisecond. Such RTT includes the
		// processing delays caused by NodeHosts, not just the network delay between
		// two NodeHost instances.
		RTTMillisecond: 200,
		// RaftAddress is used to identify the NodeHost instance
		RaftAddress: nodeAddr,
	}
	nh, err := dragonboat.NewNodeHost(nhc)
	if err != nil {
		panic(err)
	}
	if err := nh.StartCluster(initialMembers, *join, NewExampleStateMachine, rc); err != nil {
		fmt.Fprintf(os.Stderr, "failed to add cluster, %v\n", err)
		os.Exit(1)
	}
	raftStopper := syncutil.NewStopper()
	consoleStopper := syncutil.NewStopper()
	ch := make(chan string, 16)
	consoleStopper.RunWorker(func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			s, err := reader.ReadString('\n')
			if err != nil {
				close(ch)
				return
			}
			if s == "exit\n" {
				raftStopper.Stop()
				// no data will be lost/corrupted if nodehost.Stop() is not called
				nh.Stop()
				return
			}
			ch <- s
		}
	})
	raftStopper.RunWorker(func() {
		// this goroutine makes a linearizable read every 10 second. it returns the
		// Count value maintained in IStateMachine. see datastore.go for details.
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				result, err := nh.SyncRead(ctx, exampleClusterID, []byte{})
				cancel()
				if err == nil {
					var count uint64
					count = binary.LittleEndian.Uint64(result.([]byte))
					fmt.Fprintf(os.Stdout, "count: %d\n", count)
				}
			case <-raftStopper.ShouldStop():
				return
			}
		}
	})
	raftStopper.RunWorker(func() {
		// use a NO-OP client session here
		// check the example in godoc to see how to use a regular client session
		cs := nh.GetNoOPSession(exampleClusterID)
		for {
			select {
			case v, ok := <-ch:
				if !ok {
					return
				}
				// remove the \n char
				msg := strings.Replace(v, "\n", "", 1)
				if cmd, addr, nodeID, err := splitMembershipChangeCmd(msg); err == nil {
					// input is a membership change request
					makeMembershipChange(nh, cmd, addr, nodeID)
				} else {
					// input is a regular message need to be proposed
					ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					// make a proposal to update the IStateMachine instance
					_, err := nh.SyncPropose(ctx, cs, []byte(msg))
					cancel()
					if err != nil {
						fmt.Fprintf(os.Stderr, "SyncPropose returned error %v\n", err)
					}
				}
			case <-raftStopper.ShouldStop():
				return
			}
		}
	})
	raftStopper.Wait()
}

// splitMembershipChangeCmd tries to parse the input string as membership change
// request. ADD node request has the following expected format -
// add localhost:63100 4
// REMOVE node request has the following expected format -
// remove 4
func splitMembershipChangeCmd(v string) (string, string, uint64, error) {
	parts := strings.Split(v, " ")
	if len(parts) == 2 || len(parts) == 3 {
		cmd := strings.ToLower(strings.TrimSpace(parts[0]))
		if cmd != "add" && cmd != "remove" {
			return "", "", 0, errNotMembershipChange
		}
		addr := ""
		var nodeIDStr string
		var nodeID uint64
		var err error
		if cmd == "add" {
			addr = strings.TrimSpace(parts[1])
			nodeIDStr = strings.TrimSpace(parts[2])
		} else {
			nodeIDStr = strings.TrimSpace(parts[1])
		}
		if nodeID, err = strconv.ParseUint(nodeIDStr, 10, 64); err != nil {
			return "", "", 0, errNotMembershipChange
		}
		return cmd, addr, nodeID, nil
	}
	return "", "", 0, errNotMembershipChange
}

// makeMembershipChange makes membership change request.
func makeMembershipChange(nh *dragonboat.NodeHost, cmd string, addr string, nodeID uint64) {
	var rs *dragonboat.RequestState
	var err error
	if cmd == "add" {
		// orderID is ignored in standalone mode
		rs, err = nh.RequestAddNode(exampleClusterID, nodeID, addr, 0, 3*time.Second)
	} else if cmd == "remove" {
		rs, err = nh.RequestDeleteNode(exampleClusterID, nodeID, 0, 3*time.Second)
	} else {
		panic("unknown cmd")
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "membership change failed, %v\n", err)
		return
	}
	select {
	case r := <-rs.ResultC():
		if r.Completed() {
			fmt.Fprintf(os.Stdout, "membership change completed successfully\n")
		} else {
			fmt.Fprintf(os.Stderr, "membership change failed\n")
		}
	}
}
