package server

import "k8s-lx1036/k8s/storage/sunfs/pkg/util/config"

type Server interface {
	Start(cfg *config.Config) error
	Shutdown()
	// Sync will block invoker goroutine until this MetaNode shutdown.
	Sync()
}
