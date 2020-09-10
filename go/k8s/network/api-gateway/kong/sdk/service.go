package sdk

import "github.com/astaxie/beego/httplib"

// https://docs.konghq.com/2.1.x/admin-api/#service-object
type Service struct {
	Name              string   `json:"name" yaml:"name"`
	Protocol          string   `json:"protocol" yaml:"protocol"`
	Host              string   `json:"host" yaml:"host"`
	Port              int      `json:"port" yaml:"port"`
	Path              string   `json:"path" yaml:"path"`
	Retries           int      `json:"retries" yaml:"retries"`
	ConnectTimeout    int      `json:"connect_timeout" yaml:"connect_timeout"`
	WriteTimeout      int      `json:"write_timeout" yaml:"write_timeout"`
	ReadTimeout       int      `json:"read_timeout" yaml:"read_timeout"`
	Tags              []string `json:"tags" yaml:"tags"`
	ClientCertificate struct {
		Id string `json:"id" yaml:"id"`
	} `json:"client_certificate" yaml:"client_certificate"`
	TlsVerify      bool   `json:"tls_verify" yaml:"tls_verify"`
	TlsVerifyDepth string `json:"tls_verify_depth" yaml:"tls_verify_depth"`
	CreatedAt      int    `json:"created_at" yaml:"created_at"`
	UpdatedAt      int    `json:"updated_at" yaml:"updated_at"`
}
type ServiceResponse struct {
	Id string `json:"id" yaml:"id"`
	Service
}

// https://docs.konghq.com/2.1.x/admin-api/#request-body
/*type ServiceRequest struct {
	Name           string `json:"name" yaml:"name"`
	Retries        int    `json:"retries" yaml:"retries"`
	Protocol       string `json:"protocol" yaml:"protocol"`

}*/
type ServiceClient struct {
	Config *Config
}

func (client *ServiceClient) Create(service Service) {

	response := httplib.Post("/services").Body().ToJSON()

}
