package config

// Config provides containerd configuration data for the server
type Config struct {
	// Root is the path to a directory where containerd will store persistent data
	Root string `toml:"root"`
	// State is the path to a directory where containerd will store transient data
	State string `toml:"state"`

	GRPC GRPCConfig `toml:"grpc"`
}

// GRPCConfig provides GRPC configuration for the socket
type GRPCConfig struct {
	Address        string `toml:"address"`
	TCPAddress     string `toml:"tcp_address"`
	TCPTLSCA       string `toml:"tcp_tls_ca"`
	TCPTLSCert     string `toml:"tcp_tls_cert"`
	TCPTLSKey      string `toml:"tcp_tls_key"`
	UID            int    `toml:"uid"`
	GID            int    `toml:"gid"`
	MaxRecvMsgSize int    `toml:"max_recv_message_size"`
	MaxSendMsgSize int    `toml:"max_send_message_size"`
}
