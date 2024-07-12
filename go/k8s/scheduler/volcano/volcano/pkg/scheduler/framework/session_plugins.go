package framework

import "k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/api"

func (ssn *Session) AddTaskOrderFn(name string, cf api.CompareFn) {
	ssn.taskOrderFns[name] = cf
}

func (ssn *Session) AddJobOrderFn(name string, cf api.CompareFn) {
	ssn.jobOrderFns[name] = cf
}

func (ssn *Session) AddPreemptableFn(name string, cf api.EvictableFn) {
	ssn.preemptableFns[name] = cf
}

// AddJobStarvingFns add jobStarvingFns function
func (ssn *Session) AddJobStarvingFns(name string, fn api.ValidateFn) {
	ssn.jobStarvingFns[name] = fn
}

func (ssn *Session) AddJobValidFn(name string, fn api.ValidateExFn) {
	ssn.jobValidFns[name] = fn
}

// JobValid invoke jobvalid function of the plugins
func (ssn *Session) JobValid(obj interface{}) *api.ValidateResult {
	for _, tier := range ssn.Tiers {
		for _, plugin := range tier.Plugins {
			validFn, found := ssn.jobValidFns[plugin.Name]
			if !found {
				continue
			}

			if result := validFn(obj); result != nil && !result.Pass {
				return result
			}
		}
	}

	return nil
}
