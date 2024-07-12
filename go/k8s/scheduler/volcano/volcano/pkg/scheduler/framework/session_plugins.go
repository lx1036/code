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

func (ssn *Session) JobEnqueueable(obj interface{}) bool {
	var hasFound bool
	for _, tier := range ssn.Tiers {
		for _, plugin := range tier.Plugins {
			if !isEnabled(plugin.EnabledJobEnqueued) {
				continue
			}
			fn, found := ssn.jobEnqueueableFns[plugin.Name]
			if !found {
				continue
			}

			res := fn(obj)
			if res < 0 {
				return false
			}
			if res > 0 {
				hasFound = true
			}
		}
		// if plugin exists that votes permit, meanwhile other plugin votes abstention,
		// permit job to be enqueueable, do not check next tier
		if hasFound {
			return true
		}
	}

	return true
}

func (ssn *Session) JobEnqueued(obj interface{}) {
	for _, tier := range ssn.Tiers {
		for _, plugin := range tier.Plugins {
			if !isEnabled(plugin.EnabledJobEnqueued) {
				continue
			}
			fn, found := ssn.jobEnqueuedFns[plugin.Name]
			if !found {
				continue
			}

			fn(obj)
		}
	}
}

func (ssn *Session) JobOrderFn(l, r interface{}) bool {
	for _, tier := range ssn.Tiers {
		for _, plugin := range tier.Plugins {
			if !isEnabled(plugin.EnabledJobOrder) {
				continue
			}
			jof, found := ssn.jobOrderFns[plugin.Name]
			if !found {
				continue
			}
			if j := jof(l, r); j != 0 {
				return j < 0
			}
		}
	}

	// If no job order funcs, order job by CreationTimestamp first, then by UID.
	lv := l.(*api.JobInfo)
	rv := r.(*api.JobInfo)
	if lv.CreationTimestamp.Equal(&rv.CreationTimestamp) {
		return lv.UID < rv.UID
	}
	return lv.CreationTimestamp.Before(&rv.CreationTimestamp)
}

func (ssn *Session) QueueOrderFn(l, r interface{}) bool {
	for _, tier := range ssn.Tiers {
		for _, plugin := range tier.Plugins {
			if !isEnabled(plugin.EnabledQueueOrder) {
				continue
			}
			queueOrderFn, found := ssn.queueOrderFns[plugin.Name]
			if !found {
				continue
			}
			if j := queueOrderFn(l, r); j != 0 {
				return j < 0
			}
		}
	}

	// If no queue order funcs, order queue by CreationTimestamp first, then by UID.
	lv := l.(*api.QueueInfo)
	rv := r.(*api.QueueInfo)
	if lv.Queue.CreationTimestamp.Equal(&rv.Queue.CreationTimestamp) {
		return lv.UID < rv.UID
	}
	return lv.Queue.CreationTimestamp.Before(&rv.Queue.CreationTimestamp)
}

func isEnabled(enabled *bool) bool {
	return enabled != nil && *enabled
}
