package raft

import (
	"io"
	"sync"
	"time"
)

// Future is used to represent an action that may occur in the future.
type Future interface {
	// Error blocks until the future arrives and then returns the error status
	// of the future. This may be called any number of times - all calls will
	// return the same value, however is not OK to call this method twice
	// concurrently on the same Future instance.
	// Error will only return generic errors related to raft, such
	// as ErrLeadershipLost, or ErrRaftShutdown. Some operations, such as
	// ApplyLog, may also return errors from other methods.
	Error() error
}

// errorFuture is used to return a static error.
type errorFuture struct {
	err error
}

func (e errorFuture) Error() error {
	return e.err
}

// deferError can be embedded to allow a future
// to provide an error in the future.
type deferError struct {
	err        error
	errCh      chan error
	responded  bool
	ShutdownCh chan struct{}
}

func (d *deferError) init() {
	d.errCh = make(chan error, 1)
}

func (d *deferError) Error() error {
	if d.err != nil {
		// Note that when we've received a nil error, this
		// won't trigger, but the channel is closed after
		// send so we'll still return nil below.
		return d.err
	}
	if d.errCh == nil {
		panic("waiting for response on nil channel")
	}
	select {
	case d.err = <-d.errCh:
	case <-d.ShutdownCh:
		d.err = ErrRaftShutdown
	}
	return d.err
}

func (d *deferError) respond(err error) {
	if d.errCh == nil {
		return
	}
	if d.responded {
		return
	}
	d.errCh <- err
	close(d.errCh)
	d.responded = true
}

// logFuture is used to apply a log entry and waits until
// the log is considered committed.
type logFuture struct {
	deferError
	log      Log
	response interface{}
	dispatch time.Time
}

// There are several types of requests that cause a configuration entry to
// be appended to the log. These are encoded here for leaderLoop() to process.
// This is internal to a single server.
type configurationChangeFuture struct {
	logFuture
	req configurationChangeRequest
}

// verifyFuture is used to verify the current node is still
// the leader. This is to prevent a stale read.
type verifyFuture struct {
	deferError
	notifyCh   chan *verifyFuture
	quorumSize int
	votes      int
	voteLock   sync.Mutex
}

// configurationsFuture is used to retrieve the current configurations. This is
// used to allow safe access to this information outside of the main thread.
type configurationsFuture struct {
	deferError
	configurations configurations
}

// Configuration returns the latest configuration in use by Raft.
func (c *configurationsFuture) Configuration() Configuration {
	return c.configurations.latest
}

// Index returns the index of the latest configuration in use by Raft.
func (c *configurationsFuture) Index() uint64 {
	return c.configurations.latestIndex
}

// bootstrapFuture is used to attempt a live bootstrap of the cluster. See the
// Raft object's BootstrapCluster member function for more details.
type bootstrapFuture struct {
	deferError

	// configuration is the proposed bootstrap configuration to apply.
	configuration Configuration
}

// reqSnapshotFuture is used for requesting a snapshot start.
// It is only used internally.
type reqSnapshotFuture struct {
	deferError

	// snapshot details provided by the FSM runner before responding
	index    uint64
	term     uint64
	snapshot FSMSnapshot
}

// userSnapshotFuture is used for waiting on a user-triggered snapshot to
// complete.
type userSnapshotFuture struct {
	deferError

	// opener is a function used to open the snapshot. This is filled in
	// once the future returns with no error.
	opener func() (*SnapshotMeta, io.ReadCloser, error)
}

// userRestoreFuture is used for waiting on a user-triggered restore of an
// external snapshot to complete.
type userRestoreFuture struct {
	deferError

	// meta is the metadata that belongs with the snapshot.
	meta *SnapshotMeta

	// reader is the interface to read the snapshot contents from.
	reader io.Reader
}

// leadershipTransferFuture is used to track the progress of a leadership
// transfer internally.
type leadershipTransferFuture struct {
	deferError

	ID      *ServerID
	Address *ServerAddress
}

type shutdownFuture struct {
	raft *Raft
}

func (s *shutdownFuture) Error() error {
	if s.raft == nil {
		return nil
	}
	if closeable, ok := s.raft.transport.(WithClose); ok {
		closeable.Close()
	}
	return nil
}
