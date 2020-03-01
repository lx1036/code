package args

import "net"

var Holder = &holder{}

// Argument holder structure. It is private to make sure that only 1 instance can be created. It holds all
// arguments values passed to Dashboard binary.
type holder struct {
	insecurePort            int
	port                    int
	tokenTTL                int
	metricClientCheckPeriod int

	insecureBindAddress net.IP
	bindAddress         net.IP

	defaultCertDir       string
	certFile             string
	keyFile              string
	apiServerHost        string
	metricsProvider      string
	heapsterHost         string
	sidecarHost          string
	kubeConfigFile       string
	systemBanner         string
	systemBannerSeverity string
	apiLogLevel          string
	namespace            string

	authenticationMode []string

	autoGenerateCertificates  bool
	enableInsecureLogin       bool
	disableSettingsAuthorizer bool

	enableSkipLogin bool

	localeConfig string
}

// GetInsecurePort 'insecure-port' argument of Dashboard binary.
func (self *holder) GetInsecurePort() int {
	return self.insecurePort
}

// GetPort 'port' argument of Dashboard binary.
func (self *holder) GetPort() int {
	return self.port
}

// GetTokenTTL 'token-ttl' argument of Dashboard binary.
func (self *holder) GetTokenTTL() int {
	return self.tokenTTL
}

// GetMetricClientCheckPeriod 'metric-client-check-period' argument of Dashboard binary.
func (self *holder) GetMetricClientCheckPeriod() int {
	return self.metricClientCheckPeriod
}

// GetInsecureBindAddress 'insecure-bind-address' argument of Dashboard binary.
func (self *holder) GetInsecureBindAddress() net.IP {
	return self.insecureBindAddress
}

// GetBindAddress 'bind-address' argument of Dashboard binary.
func (self *holder) GetBindAddress() net.IP {
	return self.bindAddress
}

// GetDefaultCertDir 'default-cert-dir' argument of Dashboard binary.
func (self *holder) GetDefaultCertDir() string {
	return self.defaultCertDir
}

// GetApiServerHost 'apiserver-host' argument of Dashboard binary.
func (self *holder) GetApiServerHost() string {
	return self.apiServerHost
}

// GetKubeConfigFile 'kubeconfig' argument of Dashboard binary.
func (self *holder) GetKubeConfigFile() string {
	return self.kubeConfigFile
}

// GetNamespace 'namespace' argument of Dashboard binary.
func (self *holder) GetNamespace() string {
	return self.namespace
}
