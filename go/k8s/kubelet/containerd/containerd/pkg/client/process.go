package client

import (
	"time"
)

type ExitStatus struct {
	code     uint32
	exitedAt time.Time
	err      error
}
