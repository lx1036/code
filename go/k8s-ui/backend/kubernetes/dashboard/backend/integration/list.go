package integration

import "k8s-lx1036/dashboard/backend/integration/api"

// IntegrationsGetter is responsible for listing all supported integrations.
type IntegrationsGetter interface {
	// List returns list of all supported integrations.
	List() []api.Integration
}
