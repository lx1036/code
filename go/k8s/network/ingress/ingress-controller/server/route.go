package server

import (
	"crypto/tls"
	"net/url"
	"regexp"
)

// A RoutingTable contains the information needed to route a request.
type RoutingTable struct {
	certificatesByHost map[string]map[string]*tls.Certificate
	backendsByHost     map[string][]routingTableBackend
}

type routingTableBackend struct {
	pathRE *regexp.Regexp
	url    *url.URL
}

func NewRoutingTable(payload *watcher.Payload) *RoutingTable {
	rt := &RoutingTable{
		certificatesByHost: make(map[string]map[string]*tls.Certificate),
		backendsByHost:     make(map[string][]routingTableBackend),
	}
	rt.init(payload)
	return rt
}

func (rt *RoutingTable) init(payload *watcher.Payload) {

}
