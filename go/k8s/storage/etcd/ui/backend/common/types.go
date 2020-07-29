package common


type EtcdServer struct {
	Title     string         `json:"title"`
	Name      string         `json:"name"`
	Endpoints string       `json:"endpoints"`
	Username  string         `json:"username"`
	Password  string         `json:"password"`
	KeyPrefix string         `json:"key_prefix"`
	Desc      string         `json:"desc"`
	TLSEnable bool           `json:"tls_enable"` // 是否启用tls连接
	Roles     []string       `json:"roles"`      // 可访问此etcd服务的角色列表
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
	CAFile   string `json:"ca_file"`
}
