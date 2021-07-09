package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"k8s-lx1036/k8s/storage/csi/csi-lib-utils/connection"
	"k8s-lx1036/k8s/storage/csi/csi-lib-utils/metrics"
	"k8s-lx1036/k8s/storage/csi/csi-lib-utils/rpc"

	"golang.org/x/sys/unix"
	"google.golang.org/grpc"

	"k8s.io/klog/v2"
	registerapi "k8s.io/kubelet/pkg/apis/pluginregistration/v1"
)

var (
	connectionTimeout = flag.Duration("connection-timeout", 0, "The --connection-timeout flag is deprecated")
	operationTimeout  = flag.Duration("timeout", time.Second, "Timeout for waiting for communication with driver")
	csiAddress        = flag.String("csi-address", "/tmp/plugins/csi_example/csi-hostpath.sock", "Path of the CSI driver socket that the node-driver-registrar will connect to.")

	// 官方这里使用"/registration"目录，部署daemonset时也是mount到pod内这个目录，宿主机上的目录/var/lib/kubelet/plugins_registry
	// pluginRegistrationPath  = flag.String("plugin-registration-path", "/registration", "Path to Kubernetes plugin registration directory.")
	pluginRegistrationPath  = flag.String("plugin-registration-path", "/tmp/plugins_registry", "Path to Kubernetes plugin registration directory.")
	kubeletRegistrationPath = flag.String("kubelet-registration-path", "/tmp/plugins/csi_example/csi-hostpath.sock", "Path of the CSI driver socket on the Kubernetes host machine.")
	healthzPort             = flag.Int("health-port", 8081, "(deprecated) TCP port for healthz requests. Set to 0 to disable the healthz server. Only one of `--health-port` and `--http-endpoint` can be set.")
	httpEndpoint            = flag.String("http-endpoint", "", "The TCP network address where the HTTP server for diagnostics, including the health check indicating whether the registration socket exists, will listen (example: `:8080`). The default is empty string, which means the server is disabled. Only one of `--health-port` and `--http-endpoint` can be set.")
	showVersion             = flag.Bool("version", false, "Show version.")
	version                 = "1.0.0"

	// List of supported versions
	supportedVersions = []string{"1.0.0"}
)

// 实现
type registrationServer struct {
	driverName string
	endpoint   string
	version    []string
}

func (e *registrationServer) GetInfo(c context.Context, req *registerapi.InfoRequest) (*registerapi.PluginInfo, error) {
	klog.Infof("Received GetInfo call: %+v", req)
	return &registerapi.PluginInfo{
		Type:              registerapi.CSIPlugin,
		Name:              e.driverName,
		Endpoint:          e.endpoint,
		SupportedVersions: e.version,
	}, nil
}

func (e *registrationServer) NotifyRegistrationStatus(c context.Context, status *registerapi.RegistrationStatus) (*registerapi.RegistrationStatusResponse, error) {
	klog.Infof("Received NotifyRegistrationStatus call: %+v", status)
	if !status.PluginRegistered {
		errmsg := fmt.Sprintf("Registration process failed with error: %+v, restarting registration container.", status.Error)
		klog.Error(errmsg)
		return nil, errors.New(errmsg)
	}

	return &registerapi.RegistrationStatusResponse{}, nil
}

// NewregistrationServer returns an initialized registrationServer instance
func newRegistrationServer(driverName string, endpoint string, versions []string) registerapi.RegistrationServer {
	return &registrationServer{
		driverName: driverName,
		endpoint:   endpoint,
		version:    versions,
	}
}

// INFO: 最后记得删除 /data/kubernetes/var/lib/kubelet/plugins_registry/csi.sunnyfs.share.com-reg.sock socket 文件。该方式可以复用！
func removeRegSocket(csiDriverName string) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM)
	<-sigc
	socketPath := buildSocketPath(csiDriverName)
	err := os.Remove(socketPath)
	if err != nil && !os.IsNotExist(err) {
		klog.Errorf("failed to remove socket: %s with error: %+v", socketPath, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func healthzServer(socketPath string, httpEndpoint string) {
	if httpEndpoint == "" {
		klog.Infof("Skipping healthz server because HTTP endpoint is set to: %q", httpEndpoint)
		return
	}
	klog.Infof("Starting healthz server at HTTP endpoint: %v\n", httpEndpoint)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		socketExists, err := DoesSocketExist(socketPath)
		if err == nil && socketExists {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`ok`))
			klog.Infof("health check succeeded")
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			klog.Errorf("health check failed: %+v", err)
		} else if !socketExists {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("registration socket does not exist"))
			klog.Errorf("health check failed, registration socket does not exist")
		}
	})

	klog.Fatal(http.ListenAndServe(httpEndpoint, nil))
}

func Umask(mask int) (int, error) {
	return unix.Umask(mask), nil
}

func CleanupSocketFile(socketPath string) error {
	socketExists, err := DoesSocketExist(socketPath)
	if err != nil {
		return err
	}
	if socketExists {
		if err := os.Remove(socketPath); err != nil {
			return fmt.Errorf("failed to remove stale socket %s with error: %+v", socketPath, err)
		}
	}
	return nil
}

func DoesSocketExist(socketPath string) (bool, error) {
	fi, err := os.Stat(socketPath)
	if err == nil {
		if isSocket := fi.Mode()&os.ModeSocket != 0; isSocket {
			return true, nil
		}
		return false, fmt.Errorf("file exists in socketPath %s but it's not a socket.: %+v", socketPath, fi)
	}
	if !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to stat the socket %s with error: %+v", socketPath, err)
	}
	return false, nil
}

func buildSocketPath(csiDriverName string) string {
	return fmt.Sprintf("%s/%s-reg.sock", *pluginRegistrationPath, csiDriverName)
}

// debug: go run . --csi-address 127.0.0.1:10000
// 该grpc server监听在宿主机上/var/lib/kubelet/plugins_registry/${csiDriverName}-reg.sock
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
	csiMetricsMgr := metrics.NewCSIMetricsManagerForSidecar("")

	// Once https://github.com/container-storage-interface/spec/issues/159 is
	// resolved, if plugin does not support PUBLISH_UNPUBLISH_VOLUME, then we
	// can skip adding mapping to "csi.volume.kubernetes.io/nodeid" annotation.

	klog.Infof("Attempting to open a gRPC connection with: %q", *csiAddress)
	csiConn, err := connection.Connect(*csiAddress, csiMetricsMgr)
	if err != nil {
		klog.Errorf("error connecting to CSI driver: %v", err)
		os.Exit(1)
	}

	klog.Infof("Calling CSI driver to discover driver name")
	ctx, cancel := context.WithTimeout(context.Background(), *operationTimeout)
	defer cancel()

	csiDriverName, err := rpc.GetDriverName(ctx, csiConn)
	if err != nil {
		klog.Errorf("error retreiving CSI driver name: %v", err)
		os.Exit(1)
	}

	klog.Infof("CSI driver name: %q", csiDriverName)
	csiMetricsMgr.SetDriverName(csiDriverName)

	// kubelet pluginswatcher 会通过socket ${kubeletRegistrationPath}/${csiDriverName}-reg.sock,
	// 即/var/lib/kubelet/plugins_registry/csi-hostpath-reg.sock
	// 来调用该node-driver-registrar server, 所以csi-node-registrar调用顺序：
	// kubelet pluginswatcher -> csi-node-registrar server -> csi-driver server

	// When kubeletRegistrationPath is specified then driver-registrar ONLY acts
	// as gRPC server which replies to registration requests initiated by kubelet's
	// pluginswatcher infrastructure. Node labeling is done by kubelet's csi code.
	server := newRegistrationServer(csiDriverName, *kubeletRegistrationPath, supportedVersions)
	socketPath := buildSocketPath(csiDriverName)
	if err := CleanupSocketFile(socketPath); err != nil {
		klog.Errorf("%+v", err)
		os.Exit(1)
	}

	var oldmask int
	if runtime.GOOS == "linux" {
		// Default to only user accessible socket, caller can open up later if desired
		oldmask, _ = Umask(0077)
	}
	klog.Infof("Starting Registration Server at: %s\n", socketPath)
	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		klog.Errorf("failed to listen on socket: %s with error: %+v", socketPath, err)
		os.Exit(1)
	}
	if runtime.GOOS == "linux" {
		Umask(oldmask)
	}
	klog.Infof("Registration Server started at: %s\n", socketPath)

	grpcServer := grpc.NewServer()
	// Registers kubelet plugin watcher api.
	registerapi.RegisterRegistrationServer(grpcServer, server)

	go healthzServer(socketPath, addr)
	go removeRegSocket(csiDriverName)

	// Starts service
	if err := grpcServer.Serve(lis); err != nil {
		klog.Errorf("Registration Server stopped serving: %v", err)
		os.Exit(1)
	}
	// If gRPC server is gracefully shutdown, exit
	os.Exit(0)
}
