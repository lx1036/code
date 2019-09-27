
// Package caddy implements the Caddy server manager.
//
// To use this package:
//
//   1. Set the AppName and AppVersion variables.
//   2. Call LoadCaddyfile() to get the Caddyfile.
//      Pass in the name of the server type (like "http").
//      Make sure the server type's package is imported
//      (import _ "github.com/caddyserver/caddy/caddyhttp").
//   3. Call caddy.Start() to start Caddy. You get back
//      an Instance, on which you can call Restart() to
//      restart it or Stop() to stop it.
//
// You should call Wait() on your instance to wait for
// all servers to quit before your process exits.

package caddy

import (
	"log"
	"os"
	"sync"
	"time"
)

var (
	// AppName is the name of the application.
	AppName string

	// AppVersion is the version of the application.
	AppVersion string

	// Quiet mode will not show any informative output on initialization.
	Quiet bool

	// PidFile is the path to the pidfile to create.
	PidFile string

	// GracefulTimeout is the maximum duration of a graceful shutdown.
	GracefulTimeout time.Duration

	// isUpgrade will be set to true if this process
	// was started as part of an upgrade, where a parent
	// Caddy process started this one.
	isUpgrade = os.Getenv("CADDY__UPGRADE") == "1"

	// started will be set to true when the first
	// instance is started; it never gets set to
	// false after that.
	started bool

	// mu protects the variables 'isUpgrade' and 'started'.
	mu sync.Mutex
)

func init() {
	OnProcessExit = append(OnProcessExit, func() {
		if PidFile != "" {
			os.Remove(PidFile)
		}
	})
}

// Start starts Caddy with the given Caddyfile.
//
// This function blocks until all the servers are listening.
func Start(cdyfile Input) (*Instance, error) {
	inst := &Instance{serverType: cdyfile.ServerType(), wg: new(sync.WaitGroup), Storage: make(map[interface{}]interface{})}
	err := startWithListenerFds(cdyfile, inst, nil)
	if err != nil {
		return inst, err
	}
	signalSuccessToParent()
	if pidErr := writePidFile(); pidErr != nil {
		log.Printf("[ERROR] Could not write pidfile: %v", pidErr)
	}

	// Execute instantiation events
	EmitEvent(InstanceStartupEvent, inst)

	return inst, nil
}

