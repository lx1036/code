package caddy

// IsUpgrade returns true if this process is part of an upgrade
// where a parent caddy process spawned this one to upgrade
// the binary.
func IsUpgrade() bool {
	mu.Lock()
	defer mu.Unlock()
	return isUpgrade
}

// transferGob is used if this is a child process as part of
// a graceful upgrade; it is used to map listeners to their
// index in the list of inherited file descriptors. This
// variable is not safe for concurrent access.
var loadedGob transferGob

// transferGob maps bind address to index of the file descriptor
// in the Files array passed to the child process. It also contains
// the Caddyfile contents and any other state needed by the new process.
// Used only during graceful upgrades.
type transferGob struct {
	ListenerFds map[string]uintptr
	Caddyfile   Input
}
