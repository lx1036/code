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
