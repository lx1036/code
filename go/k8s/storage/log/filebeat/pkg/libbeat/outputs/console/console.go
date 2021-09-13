package console

import "k8s.io/kube-openapi/pkg/common"

type Config struct {
	Codec codec.Config `config:"codec"`

	// old pretty settings to use if no codec is configured
	Pretty bool `config:"pretty"`

	BatchSize int
}

var defaultConfig = Config{}

func NewConsoleOutput(
	_ IndexManager,
	beat Info,
	observer Observer,
	cfg *common.Config,
) (Group, error) {
	config := defaultConfig
	err := cfg.Unpack(&config)
	if err != nil {
		return Fail(err)
	}

}
