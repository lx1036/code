package raft

import "fmt"

// BootstrapCluster initializes a server's storage with the given cluster
// configuration. This should only be called at the beginning of time for the
// cluster with an identical configuration listing all Voter servers. There is
// no need to bootstrap Nonvoter and Staging servers.
//
// A cluster can only be bootstrapped once from a single participating Voter
// server. Any further attempts to bootstrap will return an error that can be
// safely ignored.
//
// One approach is to bootstrap a single server with a configuration
// listing just itself as a Voter, then invoke AddVoter() on it to add other
// servers to the cluster.
func BootstrapCluster(conf *Config, logs LogStore, stable StableStore, snaps SnapshotStore, configuration Configuration) error {
	// Validate the Raft server config.
	if err := ValidateConfig(conf); err != nil {
		return err
	}

	// Sanity check the Raft peer configuration.
	if err := checkConfiguration(configuration); err != nil {
		return err
	}

	// Make sure the cluster is in a clean state.
	hasState, err := HasExistingState(logs, stable, snaps)
	if err != nil {
		return fmt.Errorf("failed to check for existing state: %v", err)
	}
	if hasState {
		return ErrCantBootstrap
	}

	// Set current term to 1.
	if err := stable.SetUint64(keyCurrentTerm, 1); err != nil {
		return fmt.Errorf("failed to save current term: %v", err)
	}
	// Append configuration entry to log.
	entry := &Log{
		Index: 1,
		Term:  1,
		Type:  LogConfiguration,
		Data:  EncodeConfiguration(configuration),
	}
	if err := logs.StoreLog(entry); err != nil {
		return fmt.Errorf("failed to append configuration entry to log: %v", err)
	}

	return nil
}

// HasExistingState returns true if the server has any existing state (logs,
// knowledge of a current term, or any snapshots).
func HasExistingState(logs LogStore, stable StableStore, snaps SnapshotStore) (bool, error) {
	// Make sure we don't have a current term.
	currentTerm, err := stable.GetUint64(keyCurrentTerm)
	if err == nil {
		if currentTerm > 0 {
			return true, nil
		}
	} else {
		if err.Error() != "not found" {
			return false, fmt.Errorf("failed to read current term: %v", err)
		}
	}

	// Make sure we have an empty log.
	lastIndex, err := logs.LastIndex()
	if err != nil {
		return false, fmt.Errorf("failed to get last log index: %v", err)
	}
	if lastIndex > 0 {
		return true, nil
	}

	// Make sure we have no snapshots
	snapshots, err := snaps.List()
	if err != nil {
		return false, fmt.Errorf("failed to list snapshots: %v", err)
	}
	if len(snapshots) > 0 {
		return true, nil
	}

	return false, nil
}
