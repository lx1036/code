package options

import (
	"fmt"
	"os"
	"time"

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

	SchedulerNames []string
	SchedulerConf  string
	SchedulePeriod time.Duration

	DefaultQueue      string
	NodeSelector      []string // 可以配置 volcano 只工作于指定的 nodes
	NodeWorkerThreads uint32

	CacheDumpFileDir  string
	EnableCacheDumper bool

	// IgnoredCSIProvisioners contains a list of provisioners, and pod request pvc with these provisioners will
	// not be counted in pod pvc resource request and node.Allocatable, because the spec.drivers of csinode resource
	// is always null, these provisioners usually are host path csi controllers like rancher.io/local-path and hostpath.csi.k8s.io.
	IgnoredCSIProvisioners []string
}

func (s *ServerOption) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.KubeClientOptions.Master, "master", s.KubeClientOptions.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	fs.StringVar(&s.KubeClientOptions.KubeConfig, "kubeconfig", s.KubeClientOptions.KubeConfig, "Path to kubeconfig file with authorization and master location information")
	fs.Float32Var(&s.KubeClientOptions.QPS, "kube-api-qps", defaultQPS, "QPS to use while talking with kubernetes apiserver")
	fs.IntVar(&s.KubeClientOptions.Burst, "kube-api-burst", defaultBurst, "Burst to use while talking with kubernetes apiserver")

	fs.BoolVar(&s.EnableLeaderElection, "leader-elect", false,
		"Start a leader election client and gain leadership before "+
			"executing the main loop. Enable this when running replicated vc-scheduler for high availability; it is enabled by default")

	fs.StringSliceVar(&s.NodeSelector, "node-selector", nil, "volcano only work with the labeled node, like: --node-selector=volcano.sh/role:train --node-selector=volcano.sh/role:serving")

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
