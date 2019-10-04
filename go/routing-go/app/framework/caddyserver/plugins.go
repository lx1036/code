package caddy

import (
	"fmt"
	"log"
	"net"
	"sort"
	"sync"
)

var (
	// serverTypes is a map of registered server types.
	serverTypes = make(map[string]ServerType)

	// eventHooks is a map of hook name to Hook. All hooks plugins
	// must have a name.
	eventHooks = &sync.Map{}

	// caddyfileLoaders is the list of all Caddyfile loaders
	// in registration order.
	caddyfileLoaders []caddyfileLoader

	defaultCaddyfileLoader caddyfileLoader // the default loader if all else fail
	loaderUsed             caddyfileLoader // the loader that was used (relevant for reloads)
)

// Define names for the various events
const (
	StartupEvent         EventName = "startup"
	ShutdownEvent                  = "shutdown"
	CertRenewEvent                 = "certrenew"
	InstanceStartupEvent           = "instancestartup"
	InstanceRestartEvent           = "instancerestart"
)

// caddyfileLoader pairs the name of a loader to the loader.
type caddyfileLoader struct {
	name   string
	loader Loader
}

// ServerListener pairs a server to its listener and/or packetconn.
type ServerListener struct {
	server   Server
	listener net.Listener
	packet   net.PacketConn
}

// Loader is a type that can load a Caddyfile.
// It is passed the name of the server type.
// It returns an error only if something went
// wrong, not simply if there is no Caddyfile
// for this loader to load.
//
// A Loader should only load the Caddyfile if
// a certain condition or requirement is met,
// as returning a non-nil Input value along with
// another Loader will result in an error.
// In other words, loading the Caddyfile must
// be deliberate & deterministic, not haphazard.
//
// The exception is the default Caddyfile loader,
// which will be called only if no other Caddyfile
// loaders return a non-nil Input. The default
// loader may always return an Input value.
type Loader interface {
	Load(serverType string) (Input, error)
}

// DescribePlugins returns a string describing the registered plugins.
func DescribePlugins() string {
	pl := ListPlugins()

	str := "Server types:\n"
	for _, name := range pl["server_types"] {
		str += "  " + name + "\n"
	}

	str += "\nCaddyfile loaders:\n"
	for _, name := range pl["caddyfile_loaders"] {
		str += "  " + name + "\n"
	}

	if len(pl["event_hooks"]) > 0 {
		str += "\nEvent hook plugins:\n"
		for _, name := range pl["event_hooks"] {
			str += "  hook." + name + "\n"
		}
	}

	if len(pl["clustering"]) > 0 {
		str += "\nClustering plugins:\n"
		for _, name := range pl["clustering"] {
			str += "  " + name + "\n"
		}
	}

	str += "\nOther plugins:\n"
	for _, name := range pl["others"] {
		str += "  " + name + "\n"
	}

	return str
}

// ListPlugins makes a list of the registered plugins,
// keyed by plugin type.
func ListPlugins() map[string][]string {
	p := make(map[string][]string)

	// server type plugins
	for name := range serverTypes {
		p["server_types"] = append(p["server_types"], name)
	}

	// caddyfile loaders in registration order
	for _, loader := range caddyfileLoaders {
		p["caddyfile_loaders"] = append(p["caddyfile_loaders"], loader.name)
	}
	if defaultCaddyfileLoader.name != "" {
		p["caddyfile_loaders"] = append(p["caddyfile_loaders"], defaultCaddyfileLoader.name)
	}

	// List the event hook plugins
	eventHooks.Range(func(k, _ interface{}) bool {
		p["event_hooks"] = append(p["event_hooks"], k.(string))
		return true
	})

	// alphabetize the rest of the plugins
	var others []string
	for stype, stypePlugins := range plugins {
		for name := range stypePlugins {
			var s string
			if stype != "" {
				s = stype + "."
			}
			s += name
			others = append(others, s)
		}
	}

	sort.Strings(others)
	for _, name := range others {
		p["others"] = append(p["others"], name)
	}

	return p
}

// EmitEvent executes the different hooks passing the EventType as an
// argument. This is a blocking function. Hook developers should
// use 'go' keyword if they don't want to block Caddy.
func EmitEvent(event EventName, info interface{}) {
	eventHooks.Range(func(k, v interface{}) bool {
		err := v.(EventHook)(event, info)
		if err != nil {
			log.Printf("error on '%s' hook: %v", k.(string), err)
		}
		return true
	})
}

// loadCaddyfileInput iterates the registered Caddyfile loaders
// and, if needed, calls the default loader, to load a Caddyfile.
// It is an error if any of the loaders return an error or if
// more than one loader returns a Caddyfile.
func loadCaddyfileInput(serverType string) (Input, error) {
	var loadedBy string
	var caddyfileToUse Input

	for _, l := range caddyfileLoaders {
		cdyfile, err := l.loader.Load(serverType)
		if err != nil {
			return nil, fmt.Errorf("loading Caddyfile via %s: %v", l.name, err)
		}
		if cdyfile != nil {
			if caddyfileToUse != nil {
				return nil, fmt.Errorf("Caddyfile loaded multiple times; first by %s, then by %s", loadedBy, l.name)
			}
			loaderUsed = l
			caddyfileToUse = cdyfile
			loadedBy = l.name
		}
	}

	if caddyfileToUse == nil && defaultCaddyfileLoader.loader != nil {
		cdyfile, err := defaultCaddyfileLoader.loader.Load(serverType)
		if err != nil {
			return nil, err
		}
		if cdyfile != nil {
			loaderUsed = defaultCaddyfileLoader
			caddyfileToUse = cdyfile
		}
	}

	return caddyfileToUse, nil
}
