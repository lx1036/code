package errors

import "fmt"

// ErrCorruptCheckpoint error is reported when checksum does not match
var ErrCorruptCheckpoint = fmt.Errorf("checkpoint is corrupted")

// ErrCheckpointNotFound is reported when checkpoint is not found for a given key
var ErrCheckpointNotFound = fmt.Errorf("checkpoint is not found")
