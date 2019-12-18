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
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"k8s-lx1036/app/k8s/dns/caddyserver/telemetry"
	"log"
	"net"
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

	// DefaultConfigFile is the name of the configuration file that is loaded
	// by default if no other file is specified.
	DefaultConfigFile = "Caddyfile"
)

// Input represents a Caddyfile; its contents and file path
// (which should include the file name at the end of the path).
// If path does not apply (e.g. piped input) you may use
// any understandable value. The path is mainly used for logging,
// error messages, and debugging.
type Input interface {
	// Gets the Caddyfile contents
	Body() []byte

	// Gets the path to the origin file
	Path() string

	// The type of server this input is intended for
	ServerType() string
}

// Instance contains the state of servers created as a result of
// calling Start and can be used to access or control those servers.
// It is literally an instance of a server type. Instance values
// should NOT be copied. Use *Instance for safety.
type Instance struct {
	// serverType is the name of the instance's server type
	serverType string

	// caddyfileInput is the input configuration text used for this process
	caddyfileInput Input

	// wg is used to wait for all servers to shut down
	wg *sync.WaitGroup

	// context is the context created for this instance,
	// used to coordinate the setting up of the server type
	context Context

	// servers is the list of servers with their listeners
	servers []ServerListener

	// these callbacks execute when certain events occur
	OnFirstStartup  []func() error // starting, not as part of a restart
	OnStartup       []func() error // starting, even as part of a restart
	OnRestart       []func() error // before restart commences
	OnRestartFailed []func() error // if restart failed
	OnShutdown      []func() error // stopping, even as part of a restart
	OnFinalShutdown []func() error // stopping, not as part of a restart

	// storing values on an instance is preferable to
	// global state because these will get garbage-
	// collected after in-process reloads when the
	// old instances are destroyed; use StorageMu
	// to access this value safely
	Storage   map[interface{}]interface{}
	StorageMu sync.RWMutex
}

// Server is a type that can listen and serve. It supports both
// TCP and UDP, although the UDPServer interface can be used
// for more than just UDP.
//
// If the server uses TCP, it should implement TCPServer completely.
// If it uses UDP or some other protocol, it should implement
// UDPServer completely. If it uses both, both interfaces should be
// fully implemented. Any unimplemented methods should be made as
// no-ops that simply return nil values.
type Server interface {
	TCPServer
	UDPServer
}

// TCPServer is a type that can listen and serve connections.
// A TCPServer must associate with exactly zero or one net.Listeners.
type TCPServer interface {
	// Listen starts listening by creating a new listener
	// and returning it. It does not start accepting
	// connections. For UDP-only servers, this method
	// can be a no-op that returns (nil, nil).
	Listen() (net.Listener, error)

	// Serve starts serving using the provided listener.
	// Serve must start the server loop nearly immediately,
	// or at least not return any errors before the server
	// loop begins. Serve blocks indefinitely, or in other
	// words, until the server is stopped. For UDP-only
	// servers, this method can be a no-op that returns nil.
	Serve(net.Listener) error
}

// UDPServer is a type that can listen and serve packets.
// A UDPServer must associate with exactly zero or one net.PacketConns.
type UDPServer interface {
	// ListenPacket starts listening by creating a new packetconn
	// and returning it. It does not start accepting connections.
	// TCP-only servers may leave this method blank and return
	// (nil, nil).
	ListenPacket() (net.PacketConn, error)

	// ServePacket starts serving using the provided packetconn.
	// ServePacket must start the server loop nearly immediately,
	// or at least not return any errors before the server
	// loop begins. ServePacket blocks indefinitely, or in other
	// words, until the server is stopped. For TCP-only servers,
	// this method can be a no-op that returns nil.
	ServePacket(net.PacketConn) error
}

// CaddyfileInput represents a Caddyfile as input
// and is simply a convenient way to implement
// the Input interface.
type CaddyfileInput struct {
	Filepath       string
	Contents       []byte
	ServerTypeName string
}

func init() {
	OnProcessExit = append(OnProcessExit, func() {
		if PidFile != "" {
			os.Remove(PidFile)
		}
	})
}

// Body returns c.Contents.
func (c CaddyfileInput) Body() []byte { return c.Contents }

// Path returns c.Filepath.
func (c CaddyfileInput) Path() string { return c.Filepath }

// ServerType returns c.ServerType.
func (c CaddyfileInput) ServerType() string { return c.ServerTypeName }

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

// LoadCaddyfile loads a Caddyfile by calling the plugged in
// Caddyfile loader methods. An error is returned if more than
// one loader returns a non-nil Caddyfile input. If no loaders
// load a Caddyfile, the default loader is used. If no default
// loader is registered or it returns nil, the server type's
// default Caddyfile is loaded. If the server type does not
// specify any default Caddyfile value, then an empty Caddyfile
// is returned. Consequently, this function never returns a nil
// value as long as there are no errors.
func LoadCaddyfile(serverType string) (Input, error) {
	// If we are finishing an upgrade, we must obtain the Caddyfile
	// from our parent process, regardless of configured loaders.
	if IsUpgrade() {
		err := gob.NewDecoder(os.Stdin).Decode(&loadedGob)
		if err != nil {
			return nil, err
		}
		return loadedGob.Caddyfile, nil
	}

	// Ask plugged-in loaders for a Caddyfile
	cdyfile, err := loadCaddyfileInput(serverType)
	if err != nil {
		return nil, err
	}

	// Otherwise revert to default
	if cdyfile == nil {
		cdyfile = DefaultInput(serverType)
	}

	// Still nil? Geez.
	if cdyfile == nil {
		cdyfile = CaddyfileInput{ServerTypeName: serverType}
	}

	return cdyfile, nil
}

// ValidateAndExecuteDirectives will load the server blocks from cdyfile
// by parsing it, then execute the directives configured by it and store
// the resulting server blocks into inst. If justValidate is true, parse
// callbacks will not be executed between directives, since the purpose
// is only to check the input for valid syntax.
func ValidateAndExecuteDirectives(cdyfile Input, inst *Instance, justValidate bool) error {
	// If parsing only inst will be nil, create an instance for this function call only.
	if justValidate {
		inst = &Instance{serverType: cdyfile.ServerType(), wg: new(sync.WaitGroup), Storage: make(map[interface{}]interface{})}
	}

	stypeName := cdyfile.ServerType()

	stype, err := getServerType(stypeName)
	if err != nil {
		return err
	}

	inst.caddyfileInput = cdyfile

	sblocks, err := loadServerBlocks(stypeName, cdyfile.Path(), bytes.NewReader(cdyfile.Body()))
	if err != nil {
		return err
	}

	for _, sb := range sblocks {
		for dir := range sb.Tokens {
			telemetry.AppendUnique("directives", dir)
		}
	}

	inst.context = stype.NewContext(inst)
	if inst.context == nil {
		return fmt.Errorf("server type %s produced a nil Context", stypeName)
	}

	sblocks, err = inst.context.InspectServerBlocks(cdyfile.Path(), sblocks)
	if err != nil {
		return fmt.Errorf("error inspecting server blocks: %v", err)
	}

	telemetry.Set("num_server_blocks", len(sblocks))

	return executeDirectives(inst, cdyfile.Path(), stype.Directives(), sblocks, justValidate)
}

// Servers returns the ServerListeners in i.
func (i *Instance) Servers() []ServerListener {
	return i.servers
}

// Wait blocks until all of i's servers have stopped.
func (i *Instance) Wait() {
	i.wg.Wait()
}

// CaddyfileFromPipe loads the Caddyfile input from f if f is
// not interactive input. f is assumed to be a pipe or stream,
// such as os.Stdin. If f is not a pipe, no error is returned
// but the Input value will be nil. An error is only returned
// if there was an error reading the pipe, even if the length
// of what was read is 0.
func CaddyfileFromPipe(f *os.File, serverType string) (Input, error) {
	fi, err := f.Stat()
	if err == nil && fi.Mode()&os.ModeCharDevice == 0 {
		// Note that a non-nil error is not a problem. Windows
		// will not create a stdin if there is no pipe, which
		// produces an error when calling Stat(). But Unix will
		// make one either way, which is why we also check that
		// bitmask.
		// NOTE: Reading from stdin after this fails (e.g. for the let's encrypt email address) (OS X)
		confBody, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}
		return CaddyfileInput{
			Contents:       confBody,
			Filepath:       f.Name(),
			ServerTypeName: serverType,
		}, nil
	}

	// not having input from the pipe is not itself an error,
	// just means no input to return.
	return nil, nil
}
