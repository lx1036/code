package cmd

import (
	"fmt"
	"github.com/cilium/cilium/api/v1/models"
	"github.com/cilium/cilium/pkg/api"
	"github.com/cilium/cilium/pkg/loadbalancer"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/service"
)

type putServiceID struct {
	svc *service.Service
}

func NewPutServiceIDHandler(svc *service.Service) PutServiceIDHandler {
	return &putServiceID{svc: svc}
}

func (h *putServiceID) Handle(params PutServiceIDParams) middleware.Responder {
	log.WithField(logfields.Params, logfields.Repr(params)).Debug("PUT /service/{id} request")

	if params.Config.ID == 0 {
		return api.Error(PutServiceIDFailureCode, fmt.Errorf("invalid service ID 0"))
	}

	f, err := loadbalancer.NewL3n4AddrFromModel(params.Config.FrontendAddress)
	if err != nil {
		return api.Error(PutServiceIDInvalidFrontendCode, err)
	}

	frontend := loadbalancer.L3n4AddrID{
		L3n4Addr: *f,
		ID:       loadbalancer.ID(params.Config.ID),
	}
	backends := []loadbalancer.Backend{}
	for _, v := range params.Config.BackendAddresses {
		b, err := loadbalancer.NewBackendFromBackendModel(v)
		if err != nil {
			return api.Error(PutServiceIDInvalidBackendCode, err)
		}
		backends = append(backends, *b)
	}

	var svcType loadbalancer.SVCType
	switch params.Config.Flags.Type {
	case models.ServiceSpecFlagsTypeExternalIPs:
		svcType = loadbalancer.SVCTypeExternalIPs
	case models.ServiceSpecFlagsTypeNodePort:
		svcType = loadbalancer.SVCTypeNodePort
	case models.ServiceSpecFlagsTypeLoadBalancer:
		svcType = loadbalancer.SVCTypeLoadBalancer
	case models.ServiceSpecFlagsTypeHostPort:
		svcType = loadbalancer.SVCTypeHostPort
	case models.ServiceSpecFlagsTypeLocalRedirect:
		svcType = loadbalancer.SVCTypeLocalRedirect
	default:
		svcType = loadbalancer.SVCTypeClusterIP
	}

	var svcTrafficPolicy loadbalancer.SVCTrafficPolicy
	switch params.Config.Flags.TrafficPolicy {
	case models.ServiceSpecFlagsTrafficPolicyLocal:
		svcTrafficPolicy = loadbalancer.SVCTrafficPolicyLocal
	default:
		svcTrafficPolicy = loadbalancer.SVCTrafficPolicyCluster
	}

	svcHealthCheckNodePort := params.Config.Flags.HealthCheckNodePort

	var svcName, svcNamespace string
	if params.Config.Flags != nil {
		svcName = params.Config.Flags.Name
		svcNamespace = params.Config.Flags.Namespace
	}

	p := &loadbalancer.SVC{
		Name:                svcName,
		Namespace:           svcNamespace,
		Type:                svcType,
		Frontend:            frontend,
		Backends:            backends,
		TrafficPolicy:       svcTrafficPolicy,
		HealthCheckNodePort: svcHealthCheckNodePort,
	}
	created, id, err := h.svc.UpsertService(p)
	if err == nil && id != frontend.ID {
		return api.Error(PutServiceIDInvalidFrontendCode,
			fmt.Errorf("the service provided is already registered with ID %d, please use that ID instead of %d",
				id, frontend.ID))
	} else if err != nil {
		return api.Error(PutServiceIDFailureCode, err)
	} else if created {
		return NewPutServiceIDCreated()
	} else {
		return NewPutServiceIDOK()
	}
}
