package dockershim

import (
	"context"
	"fmt"

	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

// criSupportedLogDrivers are log drivers supported by native CRI integration.
var criSupportedLogDrivers = []string{"json-file"}

// IsCRISupportedLogDriver checks whether the logging driver used by docker is
// supported by native CRI integration.
func (ds *dockerService) IsCRISupportedLogDriver() (bool, error) {
	info, err := ds.client.Info()
	if err != nil {
		return false, fmt.Errorf("failed to get docker info: %v", err)
	}
	for _, driver := range criSupportedLogDrivers {
		if info.LoggingDriver == driver { // Logging Driver: json-file
			return true, nil
		}
	}
	return false, nil
}

func (ds *dockerService) ReopenContainerLog(ctx context.Context, request *runtimeapi.ReopenContainerLogRequest) (*runtimeapi.ReopenContainerLogResponse, error) {
	panic("implement me")
}
