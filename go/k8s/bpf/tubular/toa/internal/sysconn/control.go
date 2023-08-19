package sysconn

import (
	"fmt"
	"syscall"
)

// Control invokes conn.SyscallConn().Control.
func Control(conn syscall.Conn, fn func(fd int) error) error {
	raw, err := conn.SyscallConn()
	if err != nil {
		return fmt.Errorf("SyscallConn: %w", err)
	}

	var fnErr error
	err = raw.Control(func(fd uintptr) {
		fnErr = fn(int(fd))
	})
	if err != nil {
		return fmt.Errorf("Control: %w", err)
	}
	return fnErr
}
