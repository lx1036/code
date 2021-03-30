package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"k8s-lx1036/k8s/monitor/prometheus-adapter/pkg/client"
	"k8s-lx1036/k8s/monitor/prometheus-adapter/pkg/config"
	custom_provider "k8s-lx1036/k8s/monitor/prometheus-adapter/pkg/custom-provider"
	external_provider "k8s-lx1036/k8s/monitor/prometheus-adapter/pkg/external-provider"

	basecmd "github.com/kubernetes-sigs/custom-metrics-apiserver/pkg/cmd"
	"github.com/kubernetes-sigs/custom-metrics-apiserver/pkg/provider"
	"k8s.io/klog/v2"
)

type PrometheusAdapter struct {
	basecmd.AdapterBase

	// PrometheusURL is the URL describing how to connect to Prometheus.  Query parameters configure connection options.
	PrometheusURL string
	// PrometheusAuthInCluster enables using the auth details from the in-cluster kubeconfig to connect to Prometheus
	PrometheusAuthInCluster bool
	// PrometheusAuthConf is the kubeconfig file that contains auth details used to connect to Prometheus
	PrometheusAuthConf string
	// PrometheusCAFile points to the file containing the ca-root for connecting with Prometheus
	PrometheusCAFile string
	// PrometheusClientTLSCertFile points to the file containing the client TLS cert for connecting with Prometheus
	PrometheusClientTLSCertFile string
	// PrometheusClientTLSKeyFile points to the file containing the client TLS key for connecting with Prometheus
	PrometheusClientTLSKeyFile string
	// PrometheusTokenFile points to the file that contains the bearer token when connecting with Prometheus
	PrometheusTokenFile string
	// AdapterConfigFile points to the file containing the metrics discovery configuration.
	AdapterConfigFile string
	// MetricsRelistInterval is the interval at which to relist the set of available metrics
	MetricsRelistInterval time.Duration
	// MetricsMaxAge is the period to query available metrics for
	MetricsMaxAge time.Duration

	metricsConfig *config.MetricsDiscoveryConfig
}

func (cmd *PrometheusAdapter) addFlags() {
	cmd.Flags().StringVar(&cmd.PrometheusURL, "prometheus-url", cmd.PrometheusURL,
		"URL for connecting to Prometheus.")
	cmd.Flags().BoolVar(&cmd.PrometheusAuthInCluster, "prometheus-auth-incluster", cmd.PrometheusAuthInCluster,
		"use auth details from the in-cluster kubeconfig when connecting to prometheus.")
	cmd.Flags().StringVar(&cmd.PrometheusAuthConf, "prometheus-auth-config", cmd.PrometheusAuthConf,
		"kubeconfig file used to configure auth when connecting to Prometheus.")
	cmd.Flags().StringVar(&cmd.PrometheusCAFile, "prometheus-ca-file", cmd.PrometheusCAFile,
		"Optional CA file to use when connecting with Prometheus")
	cmd.Flags().StringVar(&cmd.PrometheusClientTLSCertFile, "prometheus-client-tls-cert-file", cmd.PrometheusClientTLSCertFile,
		"Optional client TLS cert file to use when connecting with Prometheus, auto-renewal is not supported")
	cmd.Flags().StringVar(&cmd.PrometheusClientTLSKeyFile, "prometheus-client-tls-key-file", cmd.PrometheusClientTLSKeyFile,
		"Optional client TLS key file to use when connecting with Prometheus, auto-renewal is not supported")
	cmd.Flags().StringVar(&cmd.PrometheusTokenFile, "prometheus-token-file", cmd.PrometheusTokenFile,
		"Optional file containing the bearer token to use when connecting with Prometheus")
	cmd.Flags().StringVar(&cmd.AdapterConfigFile, "config", cmd.AdapterConfigFile,
		"Configuration file containing details of how to transform between Prometheus metrics "+
			"and custom metrics API resources")
	cmd.Flags().DurationVar(&cmd.MetricsRelistInterval, "metrics-relist-interval", cmd.MetricsRelistInterval, ""+
		"interval at which to re-list the set of all available metrics from Prometheus")
	cmd.Flags().DurationVar(&cmd.MetricsMaxAge, "metrics-max-age", cmd.MetricsMaxAge, ""+
		"period for which to query the set of available metrics from Prometheus")
}

func (cmd *PrometheusAdapter) makePromClient() (client.Client, error) {
	baseURL, err := url.Parse(cmd.PrometheusURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Prometheus URL %q: %v", baseURL, err)
	}

	var httpClient *http.Client

	kubeconfigHTTPClient, err := makeKubeconfigHTTPClient(cmd.PrometheusAuthInCluster, cmd.PrometheusAuthConf)
	if err != nil {
		return nil, err
	}
	httpClient = kubeconfigHTTPClient
	klog.Info("successfully using in-cluster auth")

	genericPromClient := client.NewGenericAPIClient(httpClient, baseURL)
	instrumentedGenericPromClient := client.InstrumentGenericAPIClient(genericPromClient, baseURL.String())
	return client.NewClientForAPI(instrumentedGenericPromClient), nil
}

func (cmd *PrometheusAdapter) addResourceMetricsAPI(promClient client.Client) error {

}

func (cmd *PrometheusAdapter) makeProvider(promClient client.Client, stopCh <-chan struct{}) (provider.CustomMetricsProvider, error) {

	// construct the provider and start it
	cmProvider, runner := custom_provider.NewPrometheusProvider(mapper, dynClient, promClient, namers, cmd.MetricsRelistInterval, cmd.MetricsMaxAge)
	runner.RunUntil(stopCh)

	return cmProvider, nil
}

func (cmd *PrometheusAdapter) makeExternalProvider(promClient client.Client, stopCh <-chan struct{}) (provider.ExternalMetricsProvider, error) {

	// construct the provider and start it
	emProvider, runner := external_provider.NewExternalPrometheusProvider(promClient, namers, cmd.MetricsRelistInterval)
	runner.RunUntil(stopCh)

	return emProvider, nil
}
