package options

import (
	"fmt"
	"os"

	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/kube"

	"github.com/spf13/pflag"
)

const (
	defaultQPS   = 50.0
	defaultBurst = 100
)

type DecryptFunc func(c *ServerOption) error

var ServerOpts *ServerOption

type ServerOption struct {
	KubeClientOptions kube.ClientOptions

	EnableLeaderElection bool
	LockObjectNamespace  string

	CertFile   string
	KeyFile    string
	CaCertFile string
	CertData   []byte
	KeyData    []byte
	CaCertData []byte

	EnableMetrics bool
	ListenAddress string
}

func (s *ServerOption) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.KubeClientOptions.Master, "master", s.KubeClientOptions.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	fs.StringVar(&s.KubeClientOptions.KubeConfig, "kubeconfig", s.KubeClientOptions.KubeConfig, "Path to kubeconfig file with authorization and master location information")
	fs.Float32Var(&s.KubeClientOptions.QPS, "kube-api-qps", defaultQPS, "QPS to use while talking with kubernetes apiserver")
	fs.IntVar(&s.KubeClientOptions.Burst, "kube-api-burst", defaultBurst, "Burst to use while talking with kubernetes apiserver")

}

func (s *ServerOption) CheckOptionOrDie() error {
	if s.EnableLeaderElection && s.LockObjectNamespace == "" {
		return fmt.Errorf("lock-object-namespace must not be nil when LeaderElection is enabled")
	}

	return nil
}

func (s *ServerOption) ParseCAFiles(decryptFunc DecryptFunc) error {
	if err := s.readCAFiles(); err != nil {
		return err
	}

	// users can add one function to decrypt tha data by their own way if CA data is encrypted
	if decryptFunc != nil {
		return decryptFunc(s)
	}

	return nil
}

func (s *ServerOption) readCAFiles() error {
	var err error

	s.CaCertData, err = os.ReadFile(s.CaCertFile)
	if err != nil {
		return fmt.Errorf("failed to read cacert file (%s): %v", s.CaCertFile, err)
	}

	s.CertData, err = os.ReadFile(s.CertFile)
	if err != nil {
		return fmt.Errorf("failed to read cert file (%s): %v", s.CertFile, err)
	}

	s.KeyData, err = os.ReadFile(s.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to read key file (%s): %v", s.KeyFile, err)
	}

	return nil
}

func (s *ServerOption) RegisterOptions() {
	ServerOpts = s
}

func NewServerOption() *ServerOption {
	return &ServerOption{}
}
