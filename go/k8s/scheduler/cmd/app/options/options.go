package options

import (
	"fmt"
	"os"
	"time"

	"k8s-lx1036/k8s/scheduler/cmd/app/config"

	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/events"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	cliflag "k8s.io/component-base/cli/flag"
	componentbaseconfig "k8s.io/component-base/config"
	"k8s.io/kubernetes/pkg/scheduler"
)

// Options has all the params needed to run a Scheduler
type Options struct {

	// ConfigFile is the location of the scheduler server's configuration file.
	ConfigFile string
}

// Flags returns flags for a specific scheduler by section name
func (o *Options) Flags() (nfs cliflag.NamedFlagSets) {
	fs := nfs.FlagSet("misc")
	fs.StringVar(&o.ConfigFile, "config", o.ConfigFile, "The path to the configuration file.")

	return nfs
}

// Config return a scheduler config object
func (o *Options) Config() (*config.Config, error) {
	c := &config.Config{}
	if err := o.ApplyTo(c); err != nil {
		return nil, err
	}

	client, leaderElectionClient, eventClient, err := createClients(c.ComponentConfig.ClientConnection,
		c.ComponentConfig.LeaderElection.RenewDeadline.Duration)
	if err != nil {
		return nil, err
	}

	c.EventBroadcaster = events.NewEventBroadcasterAdapter(eventClient)

	// Set up leader election if enabled.
	var leaderElectionConfig *leaderelection.LeaderElectionConfig
	if c.ComponentConfig.LeaderElection.LeaderElect {
		// Use the scheduler name in the first profile to record leader election.
		coreRecorder := c.EventBroadcaster.DeprecatedNewLegacyRecorder(c.ComponentConfig.Profiles[0].SchedulerName)
		leaderElectionConfig, err = makeLeaderElectionConfig(c.ComponentConfig.LeaderElection, leaderElectionClient, coreRecorder)
		if err != nil {
			return nil, err
		}
	}

	c.Client = client
	c.InformerFactory = informers.NewSharedInformerFactory(client, 0)
	c.PodInformer = scheduler.NewPodInformer(client, 0)
	c.LeaderElection = leaderElectionConfig

	return c, nil
}

// ApplyTo applies the scheduler options to the given scheduler app configuration.
func (o *Options) ApplyTo(c *config.Config) error {
	cfg, err := loadConfigFromFile(o.ConfigFile)
	if err != nil {
		return err
	}

	c.ComponentConfig = *cfg

	return nil
}

// Validate validates all the required options.
func (o *Options) Validate() []error {
	var errs []error

	return errs
}

// NewOptions returns default scheduler app options.
func NewOptions() (*Options, error) {
	o := &Options{}

	return o, nil
}

// makeLeaderElectionConfig builds a leader election configuration. It will
// create a new resource lock associated with the configuration.
func makeLeaderElectionConfig(config componentbaseconfig.LeaderElectionConfiguration,
	client clientset.Interface, recorder record.EventRecorder) (*leaderelection.LeaderElectionConfig, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("unable to get hostname: %v", err)
	}
	// add a uniquifier so that two processes on the same host don't accidentally both become active
	id := hostname + "_" + string(uuid.NewUUID())

	rl, err := resourcelock.New(config.ResourceLock,
		config.ResourceNamespace,
		config.ResourceName,
		client.CoreV1(),
		client.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: recorder,
		})
	if err != nil {
		return nil, fmt.Errorf("couldn't create resource lock: %v", err)
	}

	return &leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: config.LeaseDuration.Duration,
		RenewDeadline: config.RenewDeadline.Duration,
		RetryPeriod:   config.RetryPeriod.Duration,
		WatchDog:      leaderelection.NewLeaderHealthzAdaptor(time.Second * 20),
		Name:          "kube-scheduler",
	}, nil
}

// GetKubernetesClient gets the client for k8s, if ~/.kube/config exists so get that config else incluster config
func createClients(config componentbaseconfig.ClientConnectionConfiguration,
	timeout time.Duration) (clientset.Interface, clientset.Interface, clientset.Interface, error) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", config.Kubeconfig)
	if err != nil {
		return nil, nil, nil, err
	}

	kubeConfig.DisableCompression = true
	kubeConfig.AcceptContentTypes = config.AcceptContentTypes
	kubeConfig.ContentType = config.ContentType
	kubeConfig.QPS = config.QPS
	kubeConfig.Burst = int(config.Burst)

	client, err := clientset.NewForConfig(restclient.AddUserAgent(kubeConfig, "dynamic-scheduler"))
	if err != nil {
		return nil, nil, nil, err
	}

	// shallow copy, do not modify the kubeConfig.Timeout.
	restConfig := *kubeConfig
	restConfig.Timeout = timeout
	leaderElectionClient, err := clientset.NewForConfig(restclient.AddUserAgent(&restConfig, "leader-election"))
	if err != nil {
		return nil, nil, nil, err
	}

	eventClient, err := clientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil, nil, nil, err
	}

	return client, leaderElectionClient, eventClient, nil
}
