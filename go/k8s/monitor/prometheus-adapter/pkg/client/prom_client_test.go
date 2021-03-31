package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func TestSeries(test *testing.T) {
	home, _ := os.UserHomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")
	promClient, err := makePromClient("http://10.218.112.247:8080", kubeconfig)
	if err != nil {
		panic(err)
	}

	start := model.Now().Add(-10 * time.Second)
	selector := `{__name__=~"^.*_queue_(length|size)$",namespace!=""}`
	series, err := promClient.Series(context.TODO(), model.Interval{Start: start, End: 0}, Selector(selector))
	if err != nil {
		panic(err)
	}

	klog.Infof("series len %d", len(series))

	for _, labelSet := range series {
		var labels []string
		for name, value := range labelSet {
			labels = append(labels, fmt.Sprintf("%s=%s", name, value))
		}

		klog.Infof("{%s}\n", strings.Join(labels, ","))
	}
}

func makePromClient(prometheusURL, kubeconfig string) (Client, error) {
	baseURL, err := url.Parse(prometheusURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Prometheus URL %q: %v", baseURL, err)
	}

	kubeconfigHTTPClient, err := makeKubeconfigHTTPClient(kubeconfig)
	if err != nil {
		return nil, err
	}

	genericPromClient := NewGenericAPIClient(kubeconfigHTTPClient, baseURL)
	instrumentedGenericPromClient := InstrumentGenericAPIClient(genericPromClient, baseURL.String())
	return NewClientForAPI(instrumentedGenericPromClient), nil
}

// makeKubeconfigHTTPClient constructs an HTTP for connecting with the given auth options.
func makeKubeconfigHTTPClient(kubeconfig string) (*http.Client, error) {
	clientConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}

	roundTripper, err := rest.TransportFor(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to construct client transport for connecting to Prometheus: %v", err)
	}

	return &http.Client{Transport: roundTripper}, nil
}
