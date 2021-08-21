package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/lni/dragonboat/v3/statemachine"
)

// SecondStateMachine is the IStateMachine implementation used in the example
// for handling all inputs not ends with "?".
// See https://github.com/lni/dragonboat/blob/master/statemachine/rstatemachine.go for
// more details of the IStateMachine interface.
type SecondStateMachine struct {
	ClusterID uint64
	NodeID    uint64
	Count     uint64
}

// Lookup performs local lookup on the SecondStateMachine instance. In this example,
// we always return the Count value as a little endian binary encoded byte
// slice.
func (s *SecondStateMachine) Lookup(query interface{}) (interface{}, error) {
	result := make([]byte, 8)
	binary.LittleEndian.PutUint64(result, s.Count)
	return result, nil
}

// Update updates the object using the specified committed raft entry.
func (s *SecondStateMachine) Update(data []byte) (statemachine.Result, error) {
	// in this example, we print out the following message for each
	// incoming update request. we also increase the counter by one to remember
	// how many updates we have applied
	s.Count++
	fmt.Printf("from SecondStateMachine.Update(), msg: %s, count:%d\n",
		string(data), s.Count)
	return statemachine.Result{Value: uint64(len(data))}, nil
}

// SaveSnapshot saves the current IStateMachine state into a snapshot using the
// specified io.Writer object.
func (s *SecondStateMachine) SaveSnapshot(w io.Writer,
	fc statemachine.ISnapshotFileCollection, done <-chan struct{}) error {
	// as shown above, the only state that can be saved is the Count variable
	// there is no external file in this IStateMachine example, we thus leave
	// the fc untouched
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, s.Count)
	_, err := w.Write(data)
	return err
}

// RecoverFromSnapshot recovers the state using the provided snapshot.
func (s *SecondStateMachine) RecoverFromSnapshot(r io.Reader,
	files []statemachine.SnapshotFile, done <-chan struct{}) error {
	// restore the Count variable, that is the only state we maintain in this
	// example, the input files is expected to be empty
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	v := binary.LittleEndian.Uint64(data)
	s.Count = v
	return nil
}

// Close closes the IStateMachine instance. There is nothing for us to cleanup
// or release as this is a pure in memory data store. Note that the Close
// method is not guaranteed to be called as node can crash at any time.
func (s *SecondStateMachine) Close() error { return nil }

// NewSecondStateMachine creates and return a new SecondStateMachine object.
func NewSecondStateMachine(clusterID uint64, nodeID uint64) statemachine.IStateMachine {
	return &SecondStateMachine{
		ClusterID: clusterID,
		NodeID:    nodeID,
		Count:     0,
	}
}
