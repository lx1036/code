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

type ResourceConfig struct {
	MaxPoolSize            int
	MinPoolSize            int
	MinENI                 int
	MaxENI                 int
	VPC                    string
	Zone                   string
	VSwitch                []string
	ENITags                map[string]string
	SecurityGroups         []string
	InstanceID             string
	AccessID               string
	AccessSecret           string
	EniCapRatio            float64
	EniCapShift            int
	VSwitchSelectionPolicy string
	EnableENITrunking      bool
	ENICapPolicy           ENICapPolicy
}
