package cmd

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/reloader/pkg/cmd/options"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/reloader/pkg/controller"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/reloader/pkg/kube"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/reloader/pkg/metrics"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/reloader/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

// NewReloaderCommand starts the reloader controller
func NewReloaderCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reloader",
		Short: "A watcher for your Kubernetes cluster",
		Run:   startReloader,
	}

	// options
	cmd.PersistentFlags().StringVar(&options.ConfigmapUpdateOnChangeAnnotation, "configmap-annotation", "configmap.reloader.stakater.com/reload", "annotation to detect changes in configmaps, specified by name")
	cmd.PersistentFlags().StringVar(&options.SecretUpdateOnChangeAnnotation, "secret-annotation", "secret.reloader.stakater.com/reload", "annotation to detect changes in secrets, specified by name")
	cmd.PersistentFlags().StringVar(&options.ReloaderAutoAnnotation, "auto-annotation", "reloader.stakater.com/auto", "annotation to detect changes in secrets")
	cmd.PersistentFlags().StringVar(&options.AutoSearchAnnotation, "auto-search-annotation", "reloader.stakater.com/search", "annotation to detect changes in configmaps or secrets tagged with special match annotation")
	cmd.PersistentFlags().StringVar(&options.SearchMatchAnnotation, "search-match-annotation", "reloader.stakater.com/match", "annotation to mark secrets or configmapts to match the search")
	cmd.PersistentFlags().StringVar(&options.LogFormat, "log-format", "json", "Log format to use (empty string for text, or JSON")
	cmd.PersistentFlags().StringSlice("resources-to-ignore", []string{}, "list of resources to ignore (valid options 'configMaps' or 'secrets')")
	cmd.PersistentFlags().StringSlice("namespaces-to-ignore", []string{}, "list of namespaces to ignore")
	return cmd
}

func configureLogging(logFormat string) error {
	switch logFormat {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		// just let the library use default on empty string.
		if logFormat != "" {
			return fmt.Errorf("unsupported logging formatter: %q", logFormat)
		}
	}
	return nil
}

func getStringSliceFromFlags(cmd *cobra.Command, flag string) ([]string, error) {
	slice, err := cmd.Flags().GetStringSlice(flag)
	if err != nil {
		return nil, err
	}

	return slice, nil
}
func getIgnoredNamespacesList(cmd *cobra.Command) (util.List, error) {
	return getStringSliceFromFlags(cmd, "namespaces-to-ignore")
}
func getIgnoredResourcesList(cmd *cobra.Command) (util.List, error) {
	ignoredResourcesList, err := getStringSliceFromFlags(cmd, "resources-to-ignore")
	if err != nil {
		return nil, err
	}

	for _, v := range ignoredResourcesList {
		if v != "configMaps" && v != "secrets" {
			return nil, fmt.Errorf("'resources-to-ignore' only accepts 'configMaps' or 'secrets', not '%s'", v)
		}
	}

	if len(ignoredResourcesList) > 1 {
		return nil, errors.New("'resources-to-ignore' only accepts 'configMaps' or 'secrets', not both")
	}

	return ignoredResourcesList, nil
}

func startReloader(cmd *cobra.Command, args []string) {
	err := configureLogging(options.LogFormat)
	if err != nil {
		logrus.Warn(err)
	}

	logrus.Info("Starting Reloader")
	currentNamespace := os.Getenv("KUBERNETES_NAMESPACE")
	if len(currentNamespace) == 0 {
		currentNamespace = metav1.NamespaceAll
		logrus.Warnf("KUBERNETES_NAMESPACE is unset, will detect changes in all namespaces.")
	}

	// create the clientset
	clientset, err := kube.GetKubernetesClient()
	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}

	ignoredResourcesList, err := getIgnoredResourcesList(cmd)
	if err != nil {
		logrus.Fatal(err)
	}

	ignoredNamespacesList, err := getIgnoredNamespacesList(cmd)
	if err != nil {
		logrus.Fatal(err)
	}

	collectors := metrics.SetupPrometheusEndpoint()

	stop := make(chan struct{})
	defer close(stop)
	for resourceType := range kube.ResourceMap {
		if ignoredResourcesList.Contains(resourceType) {
			continue
		}

		c, err := controller.NewController(clientset, resourceType, currentNamespace, ignoredNamespacesList, collectors)
		if err != nil {
			logrus.Fatalf("%s", err)
		}

		// Now let's start the controller
		logrus.Infof("Starting Controller to watch resource type: %s", resourceType)
		go c.Run(1, stop)
	}

	// Wait forever
	<-stop
}
