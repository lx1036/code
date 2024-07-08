package cache

import "k8s.io/client-go/rest"

func New(config *rest.Config, schedulerNames []string, defaultQueue string, nodeSelectors []string,
	nodeWorkers uint32, ignoredProvisioners []string) Cache {

}
