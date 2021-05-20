package config

import (
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/config"

	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/events"
	"k8s.io/client-go/tools/leaderelection"
)

// Config has all the context to run a Scheduler
type Config struct {
	// ComponentConfig is the scheduler server's configuration object.
	ComponentConfig config.KubeSchedulerConfiguration

	Client          clientset.Interface
	InformerFactory informers.SharedInformerFactory
	PodInformer     coreinformers.PodInformer

	// LeaderElection is optional.
	LeaderElection *leaderelection.LeaderElectionConfig

	EventBroadcaster events.EventBroadcasterAdapter
}
