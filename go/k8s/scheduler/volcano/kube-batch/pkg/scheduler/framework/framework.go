package framework

import (
	"k8s-lx1036/k8s/scheduler/volcano/kube-batch/pkg/scheduler/cache"
	"k8s-lx1036/k8s/scheduler/volcano/kube-batch/pkg/scheduler/conf"
)

func OpenSession(cache cache.Cache, tiers []conf.Tier, configurations []conf.Configuration) *Session {

	for _, tier := range tiers {
		for _, plugin := range tier.Plugins {

		}
	}

}

func CloseSession(ssn *Session) {

}
