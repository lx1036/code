package sdk

type Config struct {
	HostAddress        string
	Username           string
	Password           string
	InsecureSkipVerify bool
	ApiKey             string
	AdminToken         string
}

func (config *Config) Url(url string) {

}

func NewDefaultConfig(config Config) *Config {
	hostAddress := config.HostAddress
	if len(hostAddress) == 0 {
		hostAddress = "http://localhost:8001"
	}

	return &Config{
		HostAddress:        hostAddress,
		Username:           config.Username,
		Password:           config.Password,
		InsecureSkipVerify: config.InsecureSkipVerify,
		ApiKey:             config.ApiKey,
		AdminToken:         config.AdminToken,
	}
}
