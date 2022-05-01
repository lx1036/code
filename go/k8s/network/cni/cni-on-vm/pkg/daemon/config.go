package daemon

import (
	"os"

	"k8s.io/apimachinery/pkg/util/json"
)

type DaemonConfig struct {
	EnableENITrunking bool `yaml:"enable_eni_trunking" json:"enable_eni_trunking"`
}

func GetDaemonConfig(filePath string) (*DaemonConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	config := &DaemonConfig{}
	err = json.Unmarshal(data, config)
	return config, err
}

type PoolConfig struct {
}
