package healthserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cilium/cilium/pkg/counter"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"net/http"
	"sync/atomic"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/loadbalancer"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
)

var log = logging.DefaultLogger.WithField(logfields.LogSubsys, "service-healthserver")

// ServiceName is the name and namespace of the service
type ServiceName struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

// Service represents the object returned by the health server
type Service struct {
	Service        ServiceName `json:"service"`
	LocalEndpoints int         `json:"localEndpoints"`
}

// NewService creates a new service
func NewService(ns, name string, localEndpoints int) *Service {
	return &Service{
		Service: ServiceName{
			Namespace: ns,
			Name:      name,
		},
		LocalEndpoints: localEndpoints,
	}
}

// healthHTTPServer is a running HTTP health server for a certain service
type healthHTTPServer interface {
	updateService(*Service)
	shutdown()
}

// healthHTTPServerFactory creates a new HTTP health server, used for mocking
type healthHTTPServerFactory interface {
	newHTTPHealthServer(port uint16, svc *Service) healthHTTPServer
}

// ServiceHealthServer manages HTTP health check ports. For each added service,
// it opens a HTTP server on the specified HealthCheckNodePort and either
// responds with 200 OK if there are local endpoints for the service, or with
// 503 Service Unavailable if the service does not have any local endpoints.
type ServiceHealthServer struct {
	healthHTTPServerByPort  map[uint16]healthHTTPServer
	portRefCount            counter.IntCounter
	portByServiceID         map[loadbalancer.ID]uint16
	healthHTTPServerFactory healthHTTPServerFactory
}

// New creates a new health service server which services health checks by
// serving an HTTP endpoint for each service on the given HealthCheckNodePort.
func New() *ServiceHealthServer {
	return WithHealthHTTPServerFactory(&httpHealthHTTPServerFactory{})
}

// WithHealthHTTPServerFactory creates a new health server with a specific health
// server factory for testing purposes.
func WithHealthHTTPServerFactory(healthHTTPServerFactory healthHTTPServerFactory) *ServiceHealthServer {
	return &ServiceHealthServer{
		healthHTTPServerByPort:  map[uint16]healthHTTPServer{},
		portRefCount:            counter.IntCounter{},
		portByServiceID:         map[loadbalancer.ID]uint16{},
		healthHTTPServerFactory: healthHTTPServerFactory,
	}
}

type httpHealthHTTPServerFactory struct{}
type httpHealthServer struct {
	http.Server
	service atomic.Value
}

func (h *httpHealthHTTPServerFactory) newHTTPHealthServer(port uint16, svc *Service) healthHTTPServer {
	srv := &httpHealthServer{}
	srv.service.Store(svc)
	srv.Server = http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: srv,
	}

	go func() {
		log.WithFields(logrus.Fields{
			logfields.ServiceName:                svc.Service.Name,
			logfields.ServiceNamespace:           svc.Service.Namespace,
			logfields.ServiceHealthCheckNodePort: port,
		}).Debug("Starting new service health server")

		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			svc := srv.loadService()
			if errors.Is(err, unix.EADDRINUSE) {
				log.WithError(err).WithFields(logrus.Fields{
					logfields.ServiceName:                svc.Service.Name,
					logfields.ServiceNamespace:           svc.Service.Namespace,
					logfields.ServiceHealthCheckNodePort: port,
				}).Errorf("ListenAndServe failed for service health server, since the user might be running with kube-proxy. Please ensure that '--%s' option is set to false if '--%s' is set to '%s'", option.EnableHealthCheckNodePort, option.KubeProxyReplacement, option.KubeProxyReplacementPartial)
			}
			log.WithError(err).WithFields(logrus.Fields{
				logfields.ServiceName:                svc.Service.Name,
				logfields.ServiceNamespace:           svc.Service.Namespace,
				logfields.ServiceHealthCheckNodePort: port,
			}).Error("ListenAndServe failed for service health server")
		}
	}()

	return srv
}

func (h *httpHealthServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Use headers and JSON output compatible with kube-proxy
	svc := h.loadService()
	if svc.LocalEndpoints == 0 {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if err := json.NewEncoder(w).Encode(&svc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *httpHealthServer) updateService(svc *Service) {
	h.service.Store(svc)
}

func (h *httpHealthServer) shutdown() {
	h.Server.Shutdown(context.Background())
}

func (h *httpHealthServer) loadService() *Service {
	return h.service.Load().(*Service)
}
