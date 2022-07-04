package k8s

const (
	// Prefix is the common prefix for all annotations
	Prefix = "io.cilium"

	// GlobalService if set to true, marks a service to become a global
	// service
	GlobalService = Prefix + "/global-service"

	// SharedService if set to false, prevents a service from being shared,
	// the default is true if GlobalService is set, otherwise false,
	// Setting the annotation SharedService to false while setting
	// GlobalService to true allows to expose remote endpoints without
	// sharing local endpoints.
	SharedService = Prefix + "/shared-service"
)
