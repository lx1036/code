package types

const (
	// system default DNS resolver configuration
	ResolvConfDefault = "/etc/resolv.conf"

	// different container runtimes
	DockerContainerRuntime = "docker"
	RemoteContainerRuntime = "remote"

	// User visible keys for managing node allocatable enforcement on the node.
	NodeAllocatableEnforcementKey = "pods"
	SystemReservedEnforcementKey  = "system-reserved"
	KubeReservedEnforcementKey    = "kube-reserved"
	NodeAllocatableNoneKey        = "none"

	// fixed width version of time.RFC3339Nano
	RFC3339NanoFixed = "2006-01-02T15:04:05.000000000Z07:00"
	// variable width RFC3339 time format for lenient parsing of strings into timestamps
	RFC3339NanoLenient = "2006-01-02T15:04:05.999999999Z07:00"
)
