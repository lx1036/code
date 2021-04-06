package bolt

import (
	"errors"
)

var (
	ErrDatabaseNotOpen = errors.New("database not open")

	ErrTxClosed = errors.New("tx closed")
)
