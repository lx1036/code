package client

import (
	"flag"
	"fmt"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"os"
	"os/user"
	"path"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	kubeconfig, apiServerURL string
	logger                   = log.Log.WithName("controller-runtime").WithName("client").WithName("config")
)

func init() {
	// INFO: Fix this to allow double vendoring this library but still register flags on behalf of users
	flag.StringVar(&kubeconfig, "kubeconfig", "",
		"Paths to a kubeconfig. Only required if out-of-cluster.")

	// This flag is deprecated, it'll be removed in a future iteration, please switch to --kubeconfig.
	flag.StringVar(&apiServerURL, "master", "",
		"(Deprecated: switch to `--kubeconfig`) The address of the Kubernetes API server. Overrides any value in kubeconfig. "+
			"Only required if out-of-cluster.")
}

// config 加载顺序
// * --kubeconfig flag pointing at a file
//
// * KUBECONFIG environment variable pointing at a file
//
// * In-cluster config if running in cluster
//
// * $HOME/.kube/config if exists
func GetConfig() (*rest.Config, error) {
	return GetConfigWithContext("")
}
func GetConfigWithContext(context string) (*rest.Config, error) {
	cfg, err := loadConfig(context)
	if err != nil {
		return nil, err
	}

	if cfg.QPS == 0.0 {
		cfg.QPS = 20.0
		cfg.Burst = 30.0
	}

	return cfg, nil
}

var loadInClusterConfig = rest.InClusterConfig

func loadConfig(context string) (*rest.Config, error) {
	// If a flag is specified with the config location, use that
	if len(kubeconfig) > 0 {
		return loadConfigWithContext(apiServerURL, &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}, context)
	}

	// If the recommended kubeconfig env variable is not specified,
	// try the in-cluster config.
	kubeconfigPath := os.Getenv(clientcmd.RecommendedConfigPathEnvVar)
	if len(kubeconfigPath) == 0 {
		if c, err := loadInClusterConfig(); err == nil {
			return c, nil
		}
	}

	// If the recommended kubeconfig env variable is set, or there
	// is no in-cluster config, try the default recommended locations.
	//
	// NOTE: For default config file locations, upstream only checks
	// $HOME for the user's home directory, but we can also try
	// os/user.HomeDir when $HOME is unset.
	//
	// TODO(jlanford): could this be done upstream?
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if _, ok := os.LookupEnv("HOME"); !ok {
		u, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("could not get current user: %v", err)
		}
		loadingRules.Precedence = append(loadingRules.Precedence, path.Join(u.HomeDir, clientcmd.RecommendedHomeDir, clientcmd.RecommendedFileName))
	}

	return loadConfigWithContext(apiServerURL, loadingRules, context)
}
func loadConfigWithContext(apiServerURL string, loader clientcmd.ClientConfigLoader, context string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loader,
		&clientcmd.ConfigOverrides{
			ClusterInfo: clientcmdapi.Cluster{
				Server: apiServerURL,
			},
			CurrentContext: context,
		}).ClientConfig()
}

func GetConfigOrDie() *rest.Config {
	config, err := GetConfig()
	if err != nil {
		logger.Error(err, "unable to get kubeconfig")
		os.Exit(1)
	}
	return config
}
