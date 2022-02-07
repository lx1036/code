package docker

import (
	"net/http"
)

// HostCapabilities represent the capabilities of the registry
// host. This also represents the set of operations for which
// the registry host may be trusted to perform.
//
// For example pushing is a capability which should only be
// performed on an upstream source, not a mirror.
// Resolving (the process of converting a name into a digest)
// must be considered a trusted operation and only done by
// a host which is trusted (or more preferably by secure process
// which can prove the provenance of the mapping). A public
// mirror should never be trusted to do a resolve action.
//
// | Registry Type    | Pull | Resolve | Push |
// |------------------|------|---------|------|
// | Public Registry  | yes  | yes     | yes  |
// | Private Registry | yes  | yes     | yes  |
// | Public Mirror    | yes  | no      | no   |
// | Private Mirror   | yes  | yes     | no   |
type HostCapabilities uint8

const (
	// HostCapabilityPull represents the capability to fetch manifests
	// and blobs by digest
	HostCapabilityPull HostCapabilities = 1 << iota

	// HostCapabilityResolve represents the capability to fetch manifests
	// by name
	HostCapabilityResolve

	// HostCapabilityPush represents the capability to push blobs and
	// manifests
	HostCapabilityPush

	// Reserved for future capabilities (i.e. search, catalog, remove)
)

// Has checks whether the capabilities list has the provide capability
func (c HostCapabilities) Has(t HostCapabilities) bool {
	return c&t == t
}

// RegistryHost represents a complete configuration for a registry
// host, representing the capabilities, authorizations, connection
// configuration, and location.
type RegistryHost struct {
	Client       *http.Client
	Authorizer   Authorizer
	Host         string
	Scheme       string
	Path         string
	Capabilities HostCapabilities
	Header       http.Header
}

func (h RegistryHost) isProxy(refhost string) bool {
	if refhost != h.Host {
		if refhost != "docker.io" || h.Host != "registry-1.docker.io" {
			return true
		}
	}
	return false
}

// RegistryHosts fetches the registry hosts for a given namespace,
// provided by the host component of an distribution image reference.
type RegistryHosts func(string) ([]RegistryHost, error)
