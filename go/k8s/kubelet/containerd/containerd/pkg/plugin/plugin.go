package plugin

import (
	"fmt"
	"sync"
)

const (
	ServicePlugin Type = "io.containerd.service.v1"

	ContentPlugin Type = "io.containerd.content.v1"

	EventPlugin Type = "io.containerd.event.v1"

	RuntimePluginV2 Type = "io.containerd.runtime.v2"

	MetadataPlugin Type = "io.containerd.metadata.v1"

	TaskMonitorPlugin Type = "io.containerd.monitor.v1"
)

var register = struct {
	sync.RWMutex
	r []*Registration
}{}

// Type is the type of the plugin
type Type string

func (t Type) String() string { return string(t) }

// Registration contains information for registering a plugin
type Registration struct {
	// ID of the plugin
	ID string
	// Type of the plugin
	Type Type
	// Config specific to the plugin
	Config interface{}
	// Requires is a list of plugins that the registered plugin requires to be available
	Requires []Type

	// InitFn is called when initializing a plugin. The registration and
	// context are passed in. The init function may modify the registration to
	// add exports, capabilities and platform support declarations.
	InitFn func(*InitContext) (interface{}, error)
	// Disable the plugin from loading
	Disable bool
}

func (r *Registration) URI() string {
	return fmt.Sprintf("%s.%s", r.Type, r.ID)
}

func (r *Registration) Init(ic *InitContext) *Plugin {
	p, err := r.InitFn(ic)
	return &Plugin{
		Registration: r,
		Config:       ic.Config,
		//Meta:         ic.Meta,
		instance: p,
		err:      err,
	}
}

// Graph returns an ordered list of registered plugins for initialization.
// Plugins in disableList specified by id will be disabled.
func Graph() (ordered []*Registration) {
	register.RLock()
	defer register.RUnlock()

	added := map[*Registration]bool{}
	for _, r := range register.r {
		if r.Disable {
			continue
		}
		children(r, added, &ordered)
		if !added[r] {
			ordered = append(ordered, r)
			added[r] = true
		}
	}
	return ordered
}

// Register allows plugins to register
func Register(r *Registration) {
	register.Lock()
	defer register.Unlock()

	register.r = append(register.r, r)
}

func children(reg *Registration, added map[*Registration]bool, ordered *[]*Registration) {
	for _, t := range reg.Requires {
		for _, r := range register.r {
			if !r.Disable && r.URI() != reg.URI() && (t == "*" || r.Type == t) {
				children(r, added, ordered)
				if !added[r] {
					*ordered = append(*ordered, r)
					added[r] = true
				}
			}
		}
	}
}
