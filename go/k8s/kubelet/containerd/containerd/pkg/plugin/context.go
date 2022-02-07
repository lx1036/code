package plugin

import (
	"context"
	"fmt"
	"path/filepath"
)

type InitContext struct {
	Context context.Context
	Root    string
	State   string
	Config  interface{}
	Address string

	plugins *Set
}

func NewContext(ctx context.Context, r *Registration, plugins *Set, root, state string) *InitContext {
	return &InitContext{
		Context: ctx,
		Root:    filepath.Join(root, r.URI()),
		State:   filepath.Join(state, r.URI()),
		plugins: plugins,
	}
}

// GetByID returns the plugin of the given type and ID
func (i *InitContext) GetByID(t Type, id string) (interface{}, error) {
	ps, err := i.GetByType(t)
	if err != nil {
		return nil, err
	}
	p, ok := ps[id]
	if !ok {
		return nil, fmt.Errorf("no %s plugins with id %s: not found", t, id)
	}

	return p.Instance()
}

// GetByType returns all plugins with the specific type.
func (i *InitContext) GetByType(t Type) (map[string]*Plugin, error) {
	p, ok := i.plugins.byTypeAndID[t]
	if !ok {
		return nil, fmt.Errorf("no plugins registered for %s: not found", t)
	}

	return p, nil
}

// Set This maintains ordering and unique indexing over the set.
type Set struct {
	ordered     []*Plugin // order of initialization
	byTypeAndID map[Type]map[string]*Plugin
}

func NewPluginSet() *Set {
	return &Set{
		byTypeAndID: make(map[Type]map[string]*Plugin),
	}
}

func (ps *Set) Add(p *Plugin) error {
	if byID, typeok := ps.byTypeAndID[p.Registration.Type]; !typeok {
		ps.byTypeAndID[p.Registration.Type] = map[string]*Plugin{
			p.Registration.ID: p,
		}
	} else if _, idok := byID[p.Registration.ID]; !idok {
		byID[p.Registration.ID] = p
	} else {
		return fmt.Errorf("plugin %v already initialized: already exists", p.Registration.URI())
	}

	ps.ordered = append(ps.ordered, p)
	return nil
}

type Plugin struct {
	Registration *Registration // registration, as initialized
	Config       interface{}   // config, as initialized
	//Meta         *Meta

	instance interface{}
	err      error // will be set if there was an error initializing the plugin
}

// Instance returns the instance and any initialization error of the plugin
func (p *Plugin) Instance() (interface{}, error) {
	return p.instance, p.err
}
