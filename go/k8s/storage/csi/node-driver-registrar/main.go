package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"k8s.io/klog/v2"
)

var (
	connectionTimeout       = flag.Duration("connection-timeout", 0, "The --connection-timeout flag is deprecated")
	operationTimeout        = flag.Duration("timeout", time.Second, "Timeout for waiting for communication with driver")
	csiAddress              = flag.String("csi-address", "/run/csi/socket", "Path of the CSI driver socket that the node-driver-registrar will connect to.")
	pluginRegistrationPath  = flag.String("plugin-registration-path", "/registration", "Path to Kubernetes plugin registration directory.")
	kubeletRegistrationPath = flag.String("kubelet-registration-path", "", "Path of the CSI driver socket on the Kubernetes host machine.")
	healthzPort             = flag.Int("health-port", 0, "(deprecated) TCP port for healthz requests. Set to 0 to disable the healthz server. Only one of `--health-port` and `--http-endpoint` can be set.")
	httpEndpoint            = flag.String("http-endpoint", "", "The TCP network address where the HTTP server for diagnostics, including the health check indicating whether the registration socket exists, will listen (example: `:8080`). The default is empty string, which means the server is disabled. Only one of `--health-port` and `--http-endpoint` can be set.")
	showVersion             = flag.Bool("version", false, "Show version.")
	version                 = "unknown"

	// List of supported versions
	supportedVersions = []string{"1.0.0"}
)

func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")

	flag.Parse()

	if *kubeletRegistrationPath == "" {
		klog.Error("kubelet-registration-path is a required parameter")
		os.Exit(1)
	}

	if *showVersion {
		fmt.Println(os.Args[0], version)
		return
	}
	klog.Infof("Version: %s", version)

	if *healthzPort > 0 && *httpEndpoint != "" {
		klog.Error("only one of `--health-port` and `--http-endpoint` can be set.")
		os.Exit(1)
	}
	var addr string
	if *healthzPort > 0 {
		addr = ":" + strconv.Itoa(*healthzPort)
	} else {
		addr = *httpEndpoint
	}

	if *connectionTimeout != 0 {
		klog.Warning("--connection-timeout is deprecated and will have no effect")
	}

	// Unused metrics manager, necessary for connection.Connect below
	csiMetricsMgr := NewCSIMetricsManagerForSidecar("")

	// Once https://github.com/container-storage-interface/spec/issues/159 is
	// resolved, if plugin does not support PUBLISH_UNPUBLISH_VOLUME, then we
	// can skip adding mapping to "csi.volume.kubernetes.io/nodeid" annotation.

	klog.V(1).Infof("Attempting to open a gRPC connection with: %q", *csiAddress)
	csiConn, err := Connect(*csiAddress, csiMetricsMgr)
	if err != nil {
		klog.Errorf("error connecting to CSI driver: %v", err)
		os.Exit(1)
	}

	klog.V(1).Infof("Calling CSI driver to discover driver name")
	ctx, cancel := context.WithTimeout(context.Background(), *operationTimeout)
	defer cancel()

	csiDriverName, err := GetDriverName(ctx, csiConn)
	if err != nil {
		klog.Errorf("error retreiving CSI driver name: %v", err)
		os.Exit(1)
	}

	klog.V(2).Infof("CSI driver name: %q", csiDriverName)
	csiMetricsMgr.SetDriverName(csiDriverName)

	// Run forever
	nodeRegister(csiDriverName, addr)
}
