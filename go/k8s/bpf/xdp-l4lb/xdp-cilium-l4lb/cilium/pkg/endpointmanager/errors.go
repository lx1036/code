package endpointmanager

import (
	"errors"
)

var (
	ErrUnsupportedID = errors.New("unsupported IP address format")
)

type ErrInvalidPrefix struct {
	// InvalidPrefix contains the invalid prefix.
	InvalidPrefix string
}

// Error returns the string representation of the ErrInvalidPrefix.
func (e ErrInvalidPrefix) Error() string {
	return "unknown endpoint prefix '" + e.InvalidPrefix + "'"
}
