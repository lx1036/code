package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

var (
	clusterCapacityCmd = &cobra.Command{
		Use:   "cluster-capacity --kubeconfig KUBECONFIG --podspec PODSPEC",
		Short: "Cluster-capacity is used for simulating scheduling of one or multiple pods",
		Run: func(cmd *cobra.Command, args []string) {
			options := NewClusterCapacityOptions()
			err := ValidateClusterCapacityOptions(options)
			if err != nil {

			}
			err = RunClusterCapacity(options)
			if err != nil {

			}
		},
	}
)

type ClusterCapacityOptions struct {
	Kubeconfig                 string
	PodSpecFile                string
	DefaultSchedulerConfigFile string
}
type ClusterCapacityConfig struct {
	Options    *ClusterCapacityOptions
	KubeClient clientset.Interface
	Pod        *corev1.Pod
}

func NewClusterCapacityConfig(options *ClusterCapacityOptions) *ClusterCapacityConfig {
	return &ClusterCapacityConfig{
		Options: options,
	}
}
func NewClusterCapacityOptions() *ClusterCapacityOptions {
	return &ClusterCapacityOptions{}
}
func ValidateClusterCapacityOptions(options *ClusterCapacityOptions) error {
	if len(options.PodSpecFile) == 0 {
		return fmt.Errorf("pod spec file is missing")
	}
	_, present := os.LookupEnv("CC_INCLUSTER")
	if !present {
		if len(options.Kubeconfig) == 0 {
			return fmt.Errorf("kubeconfig is missing")
		}
	}
	return nil
}
func RunClusterCapacity(options *ClusterCapacityOptions) error {
	conf := NewClusterCapacityConfig(options)

	opts, err := schedulerOptions.NewOptions()
	if err != nil {
		return err
	}
	opts.ConfigFile = conf.Options.DefaultSchedulerConfigFile
	completedConfig, err := utils.InitKubeSchedulerConfiguration(opts)
	if err != nil {
		return err
	}

	var config *restclient.Config
	if len(conf.Options.Kubeconfig) != 0 {
		master, err := utils.GetMasterFromKubeConfig(conf.Options.Kubeconfig)
		if err != nil {
			return fmt.Errorf("Failed to parse kubeconfig file: %v ", err)
		}
		config, err = clientcmd.BuildConfigFromFlags(master, conf.Options.Kubeconfig)
		if err != nil {
			return fmt.Errorf("unable to build config: %v", err)
		}
	} else {
		config, err = restclient.InClusterConfig()
		if err != nil {
			return fmt.Errorf("unable to build in cluster config: %v", err)
		}
	}

	conf.KubeClient, err = clientset.NewForConfig(config)
	if err != nil {
		return err
	}
	report, err := runSimulator(conf, cc)
	if err != nil {
		return err
	}

	if err := pkg.ClusterCapacityReviewPrint(); err != nil {

	}

	return nil
}

func runSimulator(config *ClusterCapacityConfig, completedConfig *schedulerConfig.CompletedConfig) {
	clusterCapacity, err := pkg.New()
	if err != nil {
		return nil, err
	}

	err = clusterCapacity.SyncWithClient(config.KubeClient)
	if err != nil {
		return nil, err
	}

	err = clusterCapacity.Run()
	if err != nil {
		return nil, err
	}

	report := clusterCapacity.Report()

	return report, nil
}

func init() {
	clusterCapacityCmd.Flags().StringVar(&kubeConfig, "kubeconfig", "", "path to the kubeconfig file.")
	clusterCapacityCmd.Flags().StringVar(&kubeConfig, "podspec", "", "path to JSON or YAML file containing pod definition.")
	clusterCapacityCmd.Flags().StringVar(&kubeConfig, "default-config", "", "path to JSON or YAML file containing scheduler configuration.")

}
