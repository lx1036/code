package raft

import (
	"bytes"
	"fmt"
	"k8s.io/klog/v2"
	"strings"
	"sync"
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
		r.snapshotStore, r.transport, configuration)
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
				before, err = r.snapshotStore.List()
				if err != nil {
					t.Fatalf("snapshot list err: %v", err)
				}
				cfg := r.config()
				if err = RecoverCluster(&cfg, &MockFSM{}, r.logs, r.stable,
					r.snapshotStore, r.transport, configuration); err != nil {
					t.Fatalf("recover err: %v", err)
				}

				// Make sure the recovery looks right.
				var after []*SnapshotMeta
				after, err = r.snapshotStore.List()
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
				r2, err = NewRaft(&cfg, &MockFSM{}, r.logs, r.stable, r.snapshotStore, trans)
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

func TestRaftApplyConcurrently(test *testing.T) {
	cluster := NewCluster(&ClusterConfig{
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
	nums := 100
	var group sync.WaitGroup
	group.Add(nums)
	applyF := func(i int) {
		defer group.Done()
		if err := leader.Apply([]byte(fmt.Sprintf("test%d", i)), 0).Error(); err != nil {
			klog.Fatalf(fmt.Sprintf("[ERR] err: %v", err))
		}
	}
	for i := 0; i < nums; i++ { // Concurrently apply
		go applyF(i)
	}
	doneCh := make(chan struct{})
	go func() {
		group.Wait() // Wait to finish
		close(doneCh)
	}()
	select {
	case <-doneCh:
	case <-time.After(time.Second * 5):
		klog.Fatalf("timeout")
	}

	cluster.EnsureLeader(cluster.Leader().localAddr)
	cluster.EnsureFSMSame()
	cluster.EnsurePeersSame()
}

func TestRaftAutoSnapshot(test *testing.T) {
	config := DefaultConfig()
	config.TrailingLogs = 10
	config.SnapshotThreshold = 50
	config.SnapshotInterval = time.Second * 1
	cluster := NewCluster(&ClusterConfig{
		Conf: config,
		Peers: []string{
			"1/127.0.0.1:7000",
			//"2/127.0.0.1:8000",
			//"3/127.0.0.1:9000",
		},
		Bootstrap: true,
	})
	defer cluster.Close()

	time.Sleep(time.Second * 3)

	leader := cluster.Leader()
	var future Future
	for i := 0; i < 100; i++ {
		future = leader.Apply([]byte(fmt.Sprintf("test%d", i)), 0)
	}
	// Wait for the last future to apply
	if err := future.Error(); err != nil {
		test.Fatalf("err: %v", err)
	}

	// Wait for a snapshot to happen
	time.Sleep(time.Second * 10)

	// Check for snapshot
	if snaps, _ := leader.snapshotStore.List(); len(snaps) == 0 {
		test.Fatalf("should have a snapshot")
	} else {
		for _, snap := range snaps {
			klog.Infof(fmt.Sprintf("snapshot meta data:%+v", *snap))
		}
	}
}

func TestRaftUserSnapshot(test *testing.T) {
	config := DefaultConfig()
	config.TrailingLogs = 10
	config.SnapshotThreshold = 50
	config.SnapshotInterval = time.Second * 1
	cluster := NewCluster(&ClusterConfig{
		Conf: config,
		Peers: []string{
			"1/127.0.0.1:7000",
		},
		Bootstrap: true,
	})
	defer cluster.Close()

	time.Sleep(time.Second * 3)

	leader := cluster.Leader()
	if err := leader.Snapshot().Error(); err != ErrNothingNewToSnapshot {
		test.Fatalf("Request for Snapshot failed: %v", err)
	}

	klog.Infof(fmt.Sprintf("apply cmd into leader..."))
	var future Future
	for i := 0; i < 100; i++ {
		future = leader.Apply([]byte(fmt.Sprintf("test%d", i)), 0)
	}
	// Wait for the last future to apply
	if err := future.Error(); err != nil {
		test.Fatalf("err: %v", err)
	}

	klog.Infof(fmt.Sprintf("manually trigger user snapshot..."))
	if err := leader.Snapshot().Error(); err != nil {
		test.Fatalf("Request for Snapshot failed: %v", err)
	}

	// Check for snapshot
	if snaps, _ := leader.snapshotStore.List(); len(snaps) == 0 {
		test.Fatalf("should have a snapshot")
	} else {
		for _, snap := range snaps {
			klog.Infof(fmt.Sprintf("snapshot meta data:%+v", *snap))
		}
	}
}

func TestRaftSnapshotAndRestore(test *testing.T) {
	fixtures := []struct {
		description string
		offset      int
	}{
		{
			description: "0",
			offset:      0,
		},
		{
			description: "1",
			offset:      1,
		},
		{
			description: "2",
			offset:      2,
		},

		// Snapshots from the future
		{
			description: "100",
			offset:      100,
		},
		{
			description: "1000",
			offset:      1000,
		},
		{
			description: "10000",
			offset:      10000,
		},
	}
	for _, fixture := range fixtures {
		test.Run(fixture.description, func(t *testing.T) {
			config := DefaultConfig()
			config.TrailingLogs = 10
			config.SnapshotThreshold = 50
			config.SnapshotInterval = time.Second * 1
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
			if err := leader.Snapshot().Error(); err != ErrNothingNewToSnapshot {
				test.Fatalf("Request for Snapshot failed: %v", err)
			}

			klog.Infof(fmt.Sprintf("apply cmd into leader..."))
			var future Future
			for i := 0; i < 100; i++ {
				future = leader.Apply([]byte(fmt.Sprintf("test%d", i)), 0)
			}
			// Wait for the last future to apply
			if err := future.Error(); err != nil {
				test.Fatalf("err: %v", err)
			}

			klog.Infof(fmt.Sprintf("manually trigger user snapshot..."))
			snapshot := leader.Snapshot()
			if err := snapshot.Error(); err != nil {
				test.Fatalf("Request for Snapshot failed: %v", err)
			}

			// Commit some more things.
			for i := 10; i < 20; i++ {
				future = leader.Apply([]byte(fmt.Sprintf("test %d", i)), 0)
			}
			if err := future.Error(); err != nil {
				test.Fatalf("Error Apply new log entries: %v", err)
			}

			// Restore the snapshot, fix the index with the offset.
			preIndex := leader.getLastIndex()
			meta, reader, err := snapshot.Open()
			meta.Index += uint64(fixture.offset)
			if err != nil {
				test.Fatalf("Snapshot open failed: %v", err)
			}
			defer reader.Close()
			if err := leader.Restore(meta, reader, 5*time.Second); err != nil {
				test.Fatalf("Restore failed: %v", err)
			}
			// Make sure the index was updated correctly. We add 2 because we burn
			// an index to create a hole, and then we apply a no-op after the
			// restore.
			var expected uint64
			if meta.Index < preIndex {
				expected = preIndex + 2
			} else {
				expected = meta.Index + 2
			}
			lastIndex := leader.getLastIndex()
			if lastIndex != expected {
				test.Fatalf("Index was not updated correctly: %d vs. %d", lastIndex, expected)
			}

			// Ensure all the logs are the same and that we have everything that was
			// part of the original snapshot, and that the contents after were
			// reverted.
			cluster.EnsureFSMSame()
			fsm := getMockFSM(cluster.fsms[0])
			fsm.Lock()
			if len(fsm.logs) != 10 {
				test.Fatalf("Log length bad: %d", len(fsm.logs))
			}
			for i, entry := range fsm.logs {
				if bytes.Compare(entry, []byte(fmt.Sprintf("test %d", i))) != 0 {
					test.Fatalf("Log entry bad: %v", entry)
				}
			}
			fsm.Unlock()
			// Commit some more things.
			for i := 20; i < 30; i++ {
				future = leader.Apply([]byte(fmt.Sprintf("test %d", i)), 0)
			}
			if err := future.Error(); err != nil {
				test.Fatalf("Error Apply new log entries: %v", err)
			}
			cluster.EnsureFSMSame()
		})
	}
}

func TestCheckLeaderLease(test *testing.T) {
	cluster := NewCluster(&ClusterConfig{
		Peers: []string{
			"1/127.0.0.1:7000",
			"2/127.0.0.1:8000",
		},
		Bootstrap: true,
	})
	defer cluster.Close()

	time.Sleep(time.Second * 3)

	leader := cluster.Leader()
	// Wait until we have a follower
	limit := time.Now().Add(cluster.longstopTimeout)
	var followers []*Raft
	for time.Now().Before(limit) && len(followers) != 1 {
		cluster.WaitEvent(nil, cluster.conf.CommitTimeout)
		followers = cluster.GetInState(Follower)
	}
	if len(followers) != 1 {
		test.Fatalf("expected a followers: %v", followers)
	}

	// Disconnect the follower now
	follower := followers[0]
	cluster.Disconnect(follower.localAddr)

	// Watch the leaderCh
	timeout := time.After(cluster.conf.LeaderLeaseTimeout * 2)
LOOP:
	for {
		select {
		case v := <-leader.LeaderCh():
			if !v {
				break LOOP
			}
		case <-timeout:
			test.Fatalf("timeout stepping down as leader")
		}
	}

	// Ensure the last contact of the leader is non-zero
	if leader.LastContact().IsZero() {
		test.Fatalf("expected non-zero contact time")
	}

	// Should be no leaders
	if len(cluster.GetInState(Leader)) != 0 {
		test.Fatalf("expected step down")
	}

	// Verify no further contact
	last := follower.LastContact()
	time.Sleep(time.Second * 3)

	// Check that last contact has not changed
	if last != follower.LastContact() {
		test.Fatalf("unexpected further contact")
	}

	// Ensure both have cleared their leader
	if l := leader.Leader(); l != "" {
		test.Fatalf("bad: %v", l)
	}
	if l := follower.Leader(); l != "" {
		test.Fatalf("bad: %v", l)
	}
}
