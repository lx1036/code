package raft

import (
	"fmt"
	"time"
)

// Config provides any necessary configuration for the Raft server.
type Config struct {
	HeartbeatTimeout time.Duration

	ElectionTimeout time.Duration

	// CommitTimeout controls the time without an Apply() operation
	// before we heartbeat to ensure a timely commit. Due to random
	// staggering, may be delayed as much as 2x this value.
	CommitTimeout time.Duration

	// MaxAppendEntries controls the maximum number of append entries
	// to send at once. We want to strike a balance between efficiency
	// and avoiding waste if the follower is going to reject because of
	// an inconsistent log.
	MaxAppendEntries int

	// BatchApplyCh indicates whether we should buffer applyCh
	// to size MaxAppendEntries. This enables batch log commitment,
	// but breaks the timeout guarantee on Apply. Specifically,
	// a log can be added to the applyCh buffer but not actually be
	// processed until after the specified timeout.
	BatchApplyCh bool

	// If we are a member of a cluster, and RemovePeer is invoked for the
	// local node, then we forget all peers and transition into the follower state.
	// If ShutdownOnRemove is set, we additional shutdown Raft. Otherwise,
	// we can become a leader of a cluster containing only this node.
	ShutdownOnRemove bool

	// TrailingLogs controls how many logs we leave after a snapshot. This is used
	// so that we can quickly replay logs on a follower instead of being forced to
	// send an entire snapshot. The value passed here is the initial setting used.
	// This can be tuned during operation using ReloadConfig.
	TrailingLogs uint64

	// NoSnapshotRestoreOnStart controls if raft will restore a snapshot to the
	// FSM on start. This is useful if your FSM recovers from other mechanisms
	// than raft snapshotting. Snapshot metadata will still be used to initialize
	// raft's configuration and index values.
	NoSnapshotRestoreOnStart bool

	// SnapshotInterval controls how often we check if we should perform a
	// snapshot. We randomly stagger between this value and 2x this value to avoid
	// the entire cluster from performing a snapshot at once. The value passed
	// here is the initial setting used. This can be tuned during operation using
	// ReloadConfig.
	SnapshotInterval time.Duration

	// SnapshotThreshold controls how many outstanding logs there must be before
	// we perform a snapshot. This is to prevent excessive snapshotting by
	// replaying a small set of logs instead. The value passed here is the initial
	// setting used. This can be tuned during operation using ReloadConfig.
	SnapshotThreshold uint64

	// LeaderLeaseTimeout is used to control how long the "lease" lasts
	// for being the leader without being able to contact a quorum
	// of nodes. If we reach this interval without contact, we will
	// step down as leader.
	LeaderLeaseTimeout time.Duration

	// LocalID is a unique ID for this server across all time. When running with
	// ProtocolVersion < 3, you must set this to be the same as the network
	// address of your transport.
	LocalID ServerID

	// NotifyCh is used to provide a channel that will be notified of leadership
	// changes. Raft will block writing to this channel, so it should either be
	// buffered or aggressively consumed.
	NotifyCh chan<- bool

	// skipStartup allows NewRaft() to bypass all background work goroutines
	skipStartup bool
}

// ServerSuffrage determines whether a Server in a Configuration gets a vote.
type ServerSuffrage int

// Note: Don't renumber these, since the numbers are written into the log.
const (
	// Voter is a server whose vote is counted in elections and whose match index
	// is used in advancing the leader's commit index.
	Voter ServerSuffrage = iota
	// Nonvoter is a server that receives log entries but is not considered for
	// elections or commitment purposes.
	Nonvoter // learner
	// Staging is a server that acts like a nonvoter with one exception: once a
	// staging server receives enough log entries to be sufficiently caught up to
	// the leader's log, the leader will invoke a membership change to change
	// the Staging server to a Voter.
	Staging // preVote
)

func (s ServerSuffrage) String() string {
	switch s {
	case Voter:
		return "Voter"
	case Nonvoter:
		return "Nonvoter"
	case Staging:
		return "Staging"
	}
	return "ServerSuffrage"
}

// ServerID is a unique string identifying a server for all time.
type ServerID string

// ServerAddress is a network address for a server that a transport can contact.
type ServerAddress string

// Server tracks the information about a single server in a configuration.
type Server struct {
	// Suffrage determines whether the server gets a vote.
	Suffrage ServerSuffrage
	// ID is a unique string identifying this server for all time.
	ID ServerID
	// Address is its network address that a transport can contact.
	Address ServerAddress
}

// Configuration tracks which servers are in the cluster, and whether they have
// votes. This should include the local server, if it's a member of the cluster.
// The servers are listed no particular order, but each should only appear once.
// These entries are appended to the log during membership changes.
type Configuration struct {
	Servers []Server
}

// Clone makes a deep copy of a Configuration.
func (c *Configuration) Clone() (copy Configuration) {
	copy.Servers = append(copy.Servers, c.Servers...)
	return
}

// checkConfiguration tests a cluster membership configuration for common
// errors.
func checkConfiguration(configuration Configuration) error {
	idSet := make(map[ServerID]bool)
	addressSet := make(map[ServerAddress]bool)
	var voters int
	for _, server := range configuration.Servers {
		if server.ID == "" {
			return fmt.Errorf("empty ID in configuration: %v", configuration)
		}
		if server.Address == "" {
			return fmt.Errorf("empty address in configuration: %v", server)
		}
		if idSet[server.ID] {
			return fmt.Errorf("found duplicate ID in configuration: %v", server.ID)
		}
		idSet[server.ID] = true
		if addressSet[server.Address] {
			return fmt.Errorf("found duplicate address in configuration: %v", server.Address)
		}
		addressSet[server.Address] = true
		if server.Suffrage == Voter {
			voters++
		}
	}
	if voters == 0 {
		return fmt.Errorf("need at least one voter in configuration: %v", configuration)
	}
	return nil
}

// hasVote returns true if the server identified by 'id' is a Voter in the
// provided Configuration.
func hasVote(configuration Configuration, id ServerID) bool {
	for _, server := range configuration.Servers {
		if server.ID == id {
			return server.Suffrage == Voter
		}
	}
	return false
}

// configurations is state tracked on every server about its Configurations.
// Note that, per Diego's dissertation, there can be at most one uncommitted
// configuration at a time (the next configuration may not be created until the
// prior one has been committed).
//
// One downside to storing just two configurations is that if you try to take a
// snapshot when your state machine hasn't yet applied the committedIndex, we
// have no record of the configuration that would logically fit into that
// snapshot. We disallow snapshotStore in that case now. An alternative approach,
// which LogCabin uses, is to track every configuration change in the
// log.
type configurations struct {
	// committed is the latest configuration in the log/snapshot that has been
	// committed (the one with the largest index).
	committed Configuration
	// committedIndex is the log index where 'committed' was written.
	committedIndex uint64
	// latest is the latest configuration in the log/snapshot (may be committed
	// or uncommitted)
	latest Configuration
	// latestIndex is the log index where 'latest' was written.
	latestIndex uint64
}

func (c *configurations) Clone() configurations {
	return configurations{
		committed:      c.committed.Clone(),
		committedIndex: c.committedIndex,
		latest:         c.latest.Clone(),
		latestIndex:    c.latestIndex,
	}
}

// ConfigurationChangeCommand is the different ways to change the cluster
// configuration.
type ConfigurationChangeCommand uint8

const (
	// AddVoter makes a server Staging unless its Voter.
	AddVoter ConfigurationChangeCommand = iota
	// AddNonvoter makes a server Nonvoter unless its Staging or Voter.
	AddNonvoter
	// DemoteVoter makes a server Nonvoter unless its absent.
	DemoteVoter
	// RemoveServer removes a server entirely from the cluster membership.
	RemoveServer
	// Promote is created automatically by a leader; it turns a Staging server
	// into a Voter.
	Promote
)

func (c ConfigurationChangeCommand) String() string {
	switch c {
	case AddVoter:
		return "AddVoter"
	case AddNonvoter:
		return "AddNonvoter"
	case DemoteVoter:
		return "DemoteVoter"
	case RemoveServer:
		return "RemoveServer"
	case Promote:
		return "Promote"
	}
	return "ConfigurationChangeCommand"
}

// configurationChangeRequest describes a change that a leader would like to
// make to its current configuration. It's used only within a single server
// (never serialized into the log), as part of `configurationChangeFuture`.
type configurationChangeRequest struct {
	command       ConfigurationChangeCommand
	serverID      ServerID
	serverAddress ServerAddress // only present for AddStaging, AddNonvoter
	// prevIndex, if nonzero, is the index of the only configuration upon which
	// this change may be applied; if another configuration entry has been
	// added in the meantime, this request will fail.
	prevIndex uint64
}

// ValidateConfig is used to validate a sane configuration
func ValidateConfig(config *Config) error {
	if len(config.LocalID) == 0 {
		return fmt.Errorf("LocalID cannot be empty")
	}
	if config.HeartbeatTimeout < 5*time.Millisecond {
		return fmt.Errorf("HeartbeatTimeout is too low")
	}
	if config.ElectionTimeout < 5*time.Millisecond {
		return fmt.Errorf("ElectionTimeout is too low")
	}
	if config.CommitTimeout < time.Millisecond {
		return fmt.Errorf("CommitTimeout is too low")
	}
	if config.MaxAppendEntries <= 0 {
		return fmt.Errorf("MaxAppendEntries must be positive")
	}
	if config.MaxAppendEntries > 1024 {
		return fmt.Errorf("MaxAppendEntries is too large")
	}
	if config.SnapshotInterval < 5*time.Millisecond {
		return fmt.Errorf("SnapshotInterval is too low")
	}
	if config.LeaderLeaseTimeout < 5*time.Millisecond {
		return fmt.Errorf("LeaderLeaseTimeout is too low")
	}
	if config.LeaderLeaseTimeout > config.HeartbeatTimeout {
		return fmt.Errorf("LeaderLeaseTimeout cannot be larger than heartbeat timeout")
	}
	if config.ElectionTimeout < config.HeartbeatTimeout {
		return fmt.Errorf("ElectionTimeout must be equal or greater than Heartbeat Timeout")
	}
	return nil
}

func DefaultConfig() *Config {
	return &Config{
		HeartbeatTimeout:   1000 * time.Millisecond,
		ElectionTimeout:    1000 * time.Millisecond,
		CommitTimeout:      50 * time.Millisecond,
		MaxAppendEntries:   64,
		ShutdownOnRemove:   true,
		TrailingLogs:       10240,
		SnapshotInterval:   120 * time.Second,
		SnapshotThreshold:  8192,
		LeaderLeaseTimeout: 500 * time.Millisecond,
	}
}

// EncodeConfiguration serializes a Configuration using MsgPack, or panics on
// errors.
func EncodeConfiguration(configuration Configuration) []byte {
	buf, err := encodeMsgPack(configuration)
	if err != nil {
		panic(fmt.Errorf("failed to encode configuration: %v", err))
	}
	return buf.Bytes()
}

// DecodeConfiguration deserializes a Configuration using MsgPack, or panics on
// errors.
func DecodeConfiguration(buf []byte) Configuration {
	var configuration Configuration
	if err := decodeMsgPack(buf, &configuration); err != nil {
		panic(fmt.Errorf("failed to decode configuration: %v", err))
	}
	return configuration
}

// nextConfiguration generates a new Configuration from the current one and a
// configuration change request. It's split from appendConfigurationEntry so
// that it can be unit tested easily.
func nextConfiguration(current Configuration, currentIndex uint64, change configurationChangeRequest) (Configuration, error) {
	if change.prevIndex > 0 && change.prevIndex != currentIndex {
		return Configuration{}, fmt.Errorf("configuration changed since %v (latest is %v)", change.prevIndex, currentIndex)
	}

	configuration := current.Clone()
	switch change.command {
	case AddVoter:
		newServer := Server{
			Suffrage: Voter,
			ID:       change.serverID,
			Address:  change.serverAddress,
		}
		found := false
		for i, server := range configuration.Servers {
			if server.ID == change.serverID {
				if server.Suffrage == Voter {
					configuration.Servers[i].Address = change.serverAddress
				} else {
					configuration.Servers[i] = newServer
				}
				found = true
				break
			}
		}
		if !found {
			configuration.Servers = append(configuration.Servers, newServer)
		}
	}

	// Make sure we didn't do something bad like remove the last voter
	if err := checkConfiguration(configuration); err != nil {
		return Configuration{}, err
	}

	return configuration, nil
}
