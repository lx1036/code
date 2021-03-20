package profile

import (
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/config"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/v1alpha1"

	"k8s.io/client-go/tools/events"
)

// Profile is a scheduling profile.
type Profile struct {
	v1alpha1.Framework
	Recorder events.EventRecorder
	Name     string
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
