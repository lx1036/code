package dnsserver

import (
	"k8s-lx1036/k8s/dns/coredns/plugin"
	"time"
)

// Quiet mode will not show any informative output on initialization.
var Quiet bool

type Server struct {
	Addr         string
	zones        map[string]*Config
	graceTimeout time.Duration
}

func NewServer(addr string, configs []*Config) (*Server, error) {
	server := &Server{
		Addr:         addr,
		graceTimeout: time.Second * 5,
	}

	for _, config := range configs {
		var stack plugin.Handler
		for i := len(config.Plugin) - 1; i >= 0; i-- {
			stack = config.Plugin[i](stack)

			config.RegisterHandler(stack)
		}
	}

	return server, nil
}
