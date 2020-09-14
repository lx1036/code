package sdk

type Client struct {
	Config *Config
}

func (client *Client) Services() *ServiceClient {
	return &ServiceClient{
		Config: client.Config,
	}
}

func NewClient(config *Config) *Client {
	return &Client{
		Config: config,
	}
}
