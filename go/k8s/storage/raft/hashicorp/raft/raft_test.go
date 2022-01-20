package raft

import (
	"fmt"
	"k8s.io/klog/v2"
	"strings"
	"testing"
	"time"
)

func TestRaftStartStop(test *testing.T) {
	cluster := NewCluster(&ClusterConfig{
		Peers: []string{
			"1/127.0.0.1:7000",
			"2/127.0.0.1:8000",
			"3/127.0.0.1:9000",
		},
		Bootstrap: true,
	})

	// wait for leader election
	time.Sleep(time.Second * 10)
	cluster.Close()

	raft := cluster.rafts[0]
	// Everything should fail now
	if f := raft.Apply(nil, 0); f.Error() != ErrRaftShutdown {
		test.Fatalf("should be shutdown: %v", f.Error())
	}
	// Should be idempotent
	if f := raft.Shutdown(); f.Error() != nil {
		test.Fatalf("shutdown should be idempotent")
	}
}

func TestRaftLiveBootstrap(test *testing.T) {
	cluster := NewCluster(&ClusterConfig{
		Peers: []string{
			"1/127.0.0.1:7000",
			"2/127.0.0.1:8000",
			"3/127.0.0.1:9000",
		},
		Bootstrap: false,
	})
	defer cluster.Close()

	// Bootstrap one of the nodes live.
	configuration := Configuration{}
	for _, r := range cluster.rafts {
		server := Server{
			ID:      r.localID,
			Address: r.localAddr,
		}
		configuration.Servers = append(configuration.Servers, server)
	}
	boot := cluster.rafts[0].BootstrapCluster(configuration)
	if err := boot.Error(); err != nil {
		test.Fatalf("bootstrap err: %v", err)
	}

	// leader election finished
	time.Sleep(time.Second * 5)

	// Should be one leader.
	cluster.Followers()
	leader := cluster.Leader()
	cluster.EnsureLeader(leader.localAddr)

	// Should be able to apply.
	future := leader.Apply([]byte("test"), cluster.conf.CommitTimeout)
	if err := future.Error(); err != nil {
		klog.Fatalf(fmt.Sprintf("apply err: %v", err))
	}
	cluster.WaitForReplication(1)

	// Make sure the live bootstrap fails now that things are started up.
	boot = cluster.rafts[0].BootstrapCluster(configuration)
	if err := boot.Error(); err != ErrCantBootstrap {
		klog.Fatalf(fmt.Sprintf("bootstrap should have failed: %v", err))
	}
}

func TestRecoverRaftClusterNoState(test *testing.T) {
	cluster := NewCluster(&ClusterConfig{
		Peers: []string{
			"1/127.0.0.1:7000",
		},
		Bootstrap: false,
	})
	defer cluster.Close()

	r := cluster.rafts[0]
	config := r.config()
	configuration := Configuration{
		Servers: []Server{
			{
				ID:      r.localID,
				Address: r.localAddr,
			},
		},
	}
	err := RecoverCluster(&config, &MockFSM{}, r.logs, r.stable,
		r.snapshots, r.transport, configuration)
	if err == nil || !strings.Contains(err.Error(), "no initial state") {
		test.Fatalf("should have failed for no initial state: %v", err)
	}
}

func TestRecoverRaftCluster(test *testing.T) {
	snapshotThreshold := 5
	fixtures := []struct {
		description  string
		appliedIndex int
	}{
		{
			description:  "no snapshot, no trailing logs",
			appliedIndex: 0,
		},
		{
			description:  "no snapshot, some trailing logs",
			appliedIndex: snapshotThreshold - 1,
		},
		{
			description:  "snapshot, with trailing logs",
			appliedIndex: snapshotThreshold + 20,
		},
	}

	for _, fixture := range fixtures {
		test.Run(fixture.description, func(t *testing.T) {
			var err error
			config := DefaultConfig()
			config.TrailingLogs = 10
			config.SnapshotThreshold = uint64(snapshotThreshold)
			cluster := NewCluster(&ClusterConfig{
				Conf: config,
				Peers: []string{
					"1/127.0.0.1:7000",
					"2/127.0.0.1:8000",
					"3/127.0.0.1:9000",
				},
				Bootstrap: true,
			})
			defer cluster.Close()

			time.Sleep(time.Second * 3)

			leader := cluster.Leader()
			for i := 0; i < fixture.appliedIndex; i++ {
				if err := leader.Apply([]byte(fmt.Sprintf("test:%d", i)), 0).Error(); err != nil {
					klog.Fatalf(fmt.Sprintf("propose/apply log err:%v", err))
				}
			}
			// Snap the configuration.
			future := leader.GetConfiguration()
			if err = future.Error(); err != nil {
				t.Fatalf("[ERR] get configuration err: %v", err)
			}
			configuration := future.Configuration()
			// Shut down the cluster.
			for _, sec := range cluster.rafts {
				if err = sec.Shutdown().Error(); err != nil {
					t.Fatalf("[ERR] shutdown err: %v", err)
				}
			}

			// Recover the cluster. We need to replace the transport and we
			// replace the FSM so no state can carry over.
			for i, r := range cluster.rafts {
				var before []*SnapshotMeta
				before, err = r.snapshots.List()
				if err != nil {
					t.Fatalf("snapshot list err: %v", err)
				}
				cfg := r.config()
				if err = RecoverCluster(&cfg, &MockFSM{}, r.logs, r.stable,
					r.snapshots, r.transport, configuration); err != nil {
					t.Fatalf("recover err: %v", err)
				}

				// Make sure the recovery looks right.
				var after []*SnapshotMeta
				after, err = r.snapshots.List()
				if err != nil {
					t.Fatalf("snapshot list err: %v", err)
				}
				if len(after) != len(before)+1 {
					t.Fatalf("expected a new snapshot, %d vs. %d", len(before), len(after))
				}
				var first uint64
				first, err = r.logs.FirstIndex()
				if err != nil {
					t.Fatalf("first log index err: %v", err)
				}
				var last uint64
				last, err = r.logs.LastIndex()
				if err != nil {
					t.Fatalf("last log index err: %v", err)
				}
				if first != 0 || last != 0 {
					t.Fatalf("expected empty logs, got %d/%d", first, last)
				}

				// Fire up the recovered Raft instance. We have to patch
				// up the cluster state manually since this is an unusual
				// operation.
				trans := NewMemoryTransport(r.localAddr)
				var r2 *Raft
				r2, err = NewRaft(&cfg, &MockFSM{}, r.logs, r.stable, r.snapshots, trans)
				if err != nil {
					t.Fatalf("new raft err: %v", err)
				}
				cluster.rafts[i] = r2
				cluster.transports[i] = r2.transport.(*MemoryTransport)
				cluster.fsms[i] = r2.fsm.(*MockFSM)
			}
			cluster.FullyConnect()
			time.Sleep(time.Second * 3)

			// Let things settle and make sure we recovered.
			cluster.EnsureLeader(cluster.Leader().localAddr)
			cluster.EnsureFSMSame()
			cluster.EnsurePeersSame()
		})
	}

}
