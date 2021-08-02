package raft

import "time"

// Config provides any necessary configuration for the Raft server.
type Config struct {
	HeartbeatTimeout time.Duration

	ElectionTimeout time.Duration
}
