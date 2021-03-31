package main

import (
	"flag"
	"os"
	"time"

	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"
)

func main() {

	logs.InitLogs()
	defer logs.FlushLogs()

	// set up flags
	cmd := &PrometheusAdapter{
		PrometheusURL:         "https://localhost",
		MetricsRelistInterval: 10 * time.Minute,
	}
	cmd.Name = "prometheus-metrics-adapter"

	//cmd.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig(generatedopenapi.GetOpenAPIDefinitions, openapinamer.NewDefinitionNamer(apiserver.Scheme))
	//cmd.OpenAPIConfig.Info.Title = "prometheus-metrics-adapter"
	//cmd.OpenAPIConfig.Info.Version = "1.0.0"
	cmd.addFlags()
	cmd.Flags().AddGoFlagSet(flag.CommandLine) // make sure we get the klog flags
	if err := cmd.Flags().Parse(os.Args); err != nil {
		klog.Fatalf("unable to parse flags: %v", err)
	}

	// if --metrics-max-age is not set, make it equal to --metrics-relist-interval
	if cmd.MetricsMaxAge == 0*time.Second {
		cmd.MetricsMaxAge = cmd.MetricsRelistInterval
	}

	// make the prometheus client
	promClient, err := cmd.makePromClient()
	if err != nil {
		klog.Fatalf("unable to construct Prometheus client: %v", err)
	}

	// load the config
	if err := cmd.loadConfig(); err != nil {
		klog.Fatalf("unable to load metrics discovery config: %v", err)
	}

	// stop channel closed on SIGTERM and SIGINT
	stopCh := genericapiserver.SetupSignalHandler()
	// construct the provider
	customProvider, err := cmd.makeCustomProvider(promClient, stopCh)
	if err != nil {
		klog.Fatalf("unable to construct custom metrics provider: %v", err)
	}
	// attach the provider to the server, if it's needed
	if customProvider != nil {
		cmd.WithCustomMetrics(customProvider)
	}

	// construct the external provider
	externalProvider, err := cmd.makeExternalProvider(promClient, stopCh)
	if err != nil {
		klog.Fatalf("unable to construct external metrics provider: %v", err)
	}
	// attach the provider to the server, if it's needed
	if externalProvider != nil {
		cmd.WithExternalMetrics(externalProvider)
	}

	// attach resource metrics support, if it's needed
	if err := cmd.addResourceMetricsAPI(promClient); err != nil {
		klog.Fatalf("unable to install resource metrics API: %v", err)
	}

	// run the server
	if err := cmd.Run(stopCh); err != nil {
		klog.Fatalf("unable to run custom metrics adapter: %v", err)
	}
}
