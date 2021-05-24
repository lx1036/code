package profile

import (
	"fmt"

	"k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/config"
	frameworkruntime "k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/runtime"
	framework "k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/v1alpha1"

	"k8s.io/client-go/tools/events"
)

// Profile is a scheduling profile.
type Profile struct {
	framework.Framework // INFO: 每一个 profile 都对应有一个 Framework
	Recorder            events.EventRecorder
	Name                string
}

// RecorderFactory builds an EventRecorder for a given scheduler name.
type RecorderFactory func(string) events.EventRecorder

// FrameworkFactory builds a Framework for a given profile configuration.
type FrameworkFactory func(config.KubeSchedulerProfile, ...frameworkruntime.Option) (framework.Framework, error)

// NewProfile builds a Profile for the given configuration.
func NewProfile(cfg config.KubeSchedulerProfile, frameworkFact FrameworkFactory, recorderFact RecorderFactory, opts ...frameworkruntime.Option) (*Profile, error) {
	recorder := recorderFact(cfg.SchedulerName)
	opts = append(opts, frameworkruntime.WithEventRecorder(recorder), frameworkruntime.WithProfileName(cfg.SchedulerName))
	fwk, err := frameworkFact(cfg, opts...)
	if err != nil {
		return nil, err
	}
	return &Profile{
		Name:      cfg.SchedulerName,
		Framework: fwk,
		Recorder:  recorder,
	}, nil
}

// Map holds profiles indexed by scheduler name.
type Map map[string]*Profile

// HandlesSchedulerName returns whether a profile handles the given scheduler name.
func (m Map) HandlesSchedulerName(name string) bool {
	_, ok := m[name]
	return ok
}

func NewMap(cfgs []config.KubeSchedulerProfile, frameworkFact FrameworkFactory,
	recorderFact RecorderFactory, opts ...frameworkruntime.Option) (Map, error) {
	m := make(Map)

	for _, cfg := range cfgs {
		p, err := NewProfile(cfg, frameworkFact, recorderFact, opts...)
		if err != nil {
			return nil, fmt.Errorf("creating profile for scheduler name %s: %v", cfg.SchedulerName, err)
		}
		m[cfg.SchedulerName] = p
	}

	return m, nil
}
