package raft

type snapshotStatus struct {
	respErr
	stopCh chan struct{}
}

func newSnapshotStatus() *snapshotStatus {
	f := &snapshotStatus{
		stopCh: make(chan struct{}),
	}
	f.init()
	return f
}
