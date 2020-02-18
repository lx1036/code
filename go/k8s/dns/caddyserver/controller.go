package caddy

import "github.com/caddyserver/caddy/caddyfile"

// Controller is given to the setup function of directives which
// gives them access to be able to read tokens with which to
// configure themselves. It also stores state for the setup
// functions, can get the current context, and can be used to
// identify a particular server block using the Key field.
type Controller struct {
	caddyfile.Dispenser

	// The instance in which the setup is occurring
	instance *Instance

	// Key is the key from the top of the server block, usually
	// an address, hostname, or identifier of some sort.
	Key string

	// OncePerServerBlock is a function that executes f
	// exactly once per server block, no matter how many
	// hosts are associated with it. If it is the first
	// time, the function f is executed immediately
	// (not deferred) and may return an error which is
	// returned by OncePerServerBlock.
	OncePerServerBlock func(f func() error) error

	// ServerBlockIndex is the 0-based index of the
	// server block as it appeared in the input.
	ServerBlockIndex int

	// ServerBlockKeyIndex is the 0-based index of this
	// key as it appeared in the input at the head of the
	// server block.
	ServerBlockKeyIndex int

	// ServerBlockKeys is a list of keys that are
	// associated with this server block. All these
	// keys, consequently, share the same tokens.
	ServerBlockKeys []string

	// ServerBlockStorage is used by a directive's
	// setup function to persist state between all
	// the keys on a server block.
	ServerBlockStorage interface{}
}
