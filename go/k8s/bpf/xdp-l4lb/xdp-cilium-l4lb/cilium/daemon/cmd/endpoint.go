package cmd

import (
    "context"
    "errors"
    "fmt"
    "github.com/cilium/cilium/pkg/annotation"
    "github.com/cilium/cilium/pkg/bandwidth"
    "github.com/cilium/cilium/pkg/k8s"
    "github.com/sirupsen/logrus"
    "net"
    "net/http"
    "runtime"
    "sync"

    "github.com/go-openapi/runtime/middleware"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/api/v1/models"
    . "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/api/v1/server/restapi/endpoint"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/api"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/endpoint"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/endpoint/regeneration"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/labels"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"
)

var errEndpointNotFound = errors.New("endpoint not found")

type getEndpoint struct {
    d *Daemon
}

func NewGetEndpointHandler(d *Daemon) GetEndpointHandler {
    return &getEndpoint{d: d}
}

func (h *getEndpoint) Handle(params GetEndpointParams) middleware.Responder {
    log.WithField(logfields.Params, fmt.Sprintf("%+v", params)).Debug("GET /endpoint request")

    // api limit
    r, err := h.d.apiLimiterSet.Wait(params.HTTPRequest.Context(), apiRequestEndpointList)
    if err != nil {
        return api.Error(http.StatusTooManyRequests, err)
    }
    defer r.Done()

    resEPs := h.d.getEndpointList(params)
    if params.Labels != nil && len(resEPs) == 0 {
        r.Error(errEndpointNotFound)
        return NewGetEndpointNotFound()
    }

    return NewGetEndpointOK().WithPayload(resEPs)
}

func (d *Daemon) getEndpointList(params GetEndpointParams) []*models.Endpoint {
    var (
        epWorkersWg, epsAppendWg sync.WaitGroup
        convertedLabels          labels.Labels
        resEPs                   []*models.Endpoint
    )

    maxGoroutines := runtime.NumCPU()
    eps := d.endpointManager.GetEndpoints()
    if len(eps) < maxGoroutines {
        maxGoroutines = len(eps)
    }

    // INFO: 为了性能，NumCPU of goroutines 并发查询. endpoints 数量很多.
    epsCh := make(chan *endpoint.Endpoint, maxGoroutines)
    epModelsCh := make(chan *models.Endpoint, maxGoroutines)

    epWorkersWg.Add(maxGoroutines)
    for i := 0; i < maxGoroutines; i++ {
        go func(wg *sync.WaitGroup, epModelsChan chan<- *models.Endpoint, epsChan <-chan *endpoint.Endpoint) {
            for ep := range epsChan {
                // params.Labels query
                if ep.HasLabels(convertedLabels) {
                    epModelsChan <- ep.GetModel()
                }
            }
            wg.Done()
        }(&epWorkersWg, epModelsCh, epsCh)
    }

    go func(epsChan chan<- *endpoint.Endpoint, eps []*endpoint.Endpoint) {
        for _, ep := range eps {
            epsChan <- ep
        }
        close(epsChan)
    }(epsCh, eps)

    epsAppendWg.Add(1)
    go func(epsAppended *sync.WaitGroup) {
        for ep := range epModelsCh {
            resEPs = append(resEPs, ep)
        }
        epsAppended.Done()
    }(&epsAppendWg)

    epWorkersWg.Wait()
    close(epModelsCh)
    epsAppendWg.Wait()

    return resEPs
}

type getEndpointID struct {
    d *Daemon
}

func NewGetEndpointIDHandler(d *Daemon) GetEndpointIDHandler {
    return &getEndpointID{d: d}
}

func (h *getEndpointID) Handle(params GetEndpointIDParams) middleware.Responder {
    log.WithField(logfields.EndpointID, params.ID).Debug("GET /endpoint/{id} request")

    r, err := h.d.apiLimiterSet.Wait(params.HTTPRequest.Context(), apiRequestEndpointGet)
    if err != nil {
        return api.Error(http.StatusTooManyRequests, err)
    }
    defer r.Done()

    ep, err := h.d.endpointManager.Lookup(params.ID)
    if err != nil {
        r.Error(err)
        return api.Error(GetEndpointIDInvalidCode, err)
    }

    if ep == nil {
        r.Error(errEndpointNotFound)
        return NewGetEndpointIDNotFound()
    }

    return NewGetEndpointIDOK().WithPayload(ep.GetModel())
}

type putEndpointID struct {
    d *Daemon
}

func NewPutEndpointIDHandler(d *Daemon) PutEndpointIDHandler {
    return &putEndpointID{d: d}
}

// Handle cni 每次创建一个 pod，都会调用该方法创建对应的 endpoint
func (h *putEndpointID) Handle(params PutEndpointIDParams) (resp middleware.Responder) {
    epTemplate := params.Endpoint
    if ep := epTemplate; ep != nil {
        log.WithField("endpoint", fmt.Sprintf("%+v", *ep)).Debug("PUT /endpoint/{id} request")
    } else {
        log.WithField(logfields.Params, fmt.Sprintf("%+v", params)).Debug("PUT /endpoint/{id} request")
        epTemplate = &models.EndpointChangeRequest{}
    }

    r, err := h.d.apiLimiterSet.Wait(params.HTTPRequest.Context(), apiRequestEndpointCreate)
    if err != nil {
        return api.Error(http.StatusTooManyRequests, err)
    }
    defer r.Done()

    ep, code, err := h.d.createEndpoint(params.HTTPRequest.Context(), h.d, epTemplate)
    if err != nil {
        r.Error(err)
        return api.Error(code, err)
    }

    ep.Logger(daemonSubsys).Info("Successful endpoint creation")

    return NewPutEndpointIDCreated()
}

func (d *Daemon) createEndpoint(ctx context.Context, owner regeneration.Owner, epTemplate *models.EndpointChangeRequest) (*endpoint.Endpoint, int, error) {
    if option.Config.EnableEndpointRoutes {
        if epTemplate.DatapathConfiguration == nil {
            epTemplate.DatapathConfiguration = &models.EndpointDatapathConfiguration{}
        }

        // Indicate to insert a per endpoint route instead of routing
        // via cilium_host interface
        epTemplate.DatapathConfiguration.InstallEndpointRoute = true

        // Since routing occurs via endpoint interface directly, BPF
        // program is needed on that device at egress as BPF program on
        // cilium_host interface is bypassed
        epTemplate.DatapathConfiguration.RequireEgressProg = true

        // Delegate routing to the Linux stack rather than tail-calling
        // between BPF programs.
        disabled := false
        epTemplate.DatapathConfiguration.RequireRouting = &disabled
    }

    log.WithFields(logrus.Fields{
        "addressing":            epTemplate.Addressing,
        logfields.ContainerID:   epTemplate.ContainerID,
        "datapathConfiguration": epTemplate.DatapathConfiguration,
        logfields.Interface:     epTemplate.InterfaceName,
        logfields.K8sPodName:    epTemplate.K8sNamespace + "/" + epTemplate.K8sPodName,
        logfields.Labels:        epTemplate.Labels,
        "sync-build":            epTemplate.SyncBuildEndpoint,
    }).Info("Create endpoint request")

    ep, err := endpoint.NewEndpointFromChangeModel(d.ctx, owner, d, d.ipcache, d.l7Proxy, d.identityAllocator, epTemplate)
    if err != nil {
        return invalidDataError(ep, fmt.Errorf("unable to parse endpoint parameters: %s", err))
    }
    oldEp := d.endpointManager.LookupCiliumID(ep.ID)
    if oldEp != nil {
        return invalidDataError(ep, fmt.Errorf("endpoint ID %d already exists", ep.ID))
    }
    oldEp = d.endpointManager.LookupContainerID(ep.GetContainerID())
    if oldEp != nil {
        return invalidDataError(ep, fmt.Errorf("endpoint for container %s already exists", ep.GetContainerID()))
    }

    // e.ID assigned here
    err = d.endpointManager.AddEndpoint(owner, ep, "Create endpoint from API PUT")
    if err != nil {
        return d.errorDuringCreation(ep, fmt.Errorf("unable to insert endpoint into manager: %s", err))
    }

    // We need to update the the visibility policy after adding the endpoint in
    // the endpoint manager because the endpoint manager create the endpoint
    // queue of the endpoint. If we execute this function before the endpoint
    // manager creates the endpoint queue the operation will fail.
    if ep.K8sNamespaceAndPodNameIsSet() && k8s.IsEnabled() && k8sLabelsConfigured {
        ep.UpdateVisibilityPolicy(func(ns, podName string) (proxyVisibility string, err error) {
            p, err := d.k8sWatcher.GetCachedPod(ns, podName)
            if err != nil {
                return "", err
            }
            return p.Annotations[annotation.ProxyVisibility], nil
        })
        ep.UpdateBandwidthPolicy(func(ns, podName string) (bandwidthEgress string, err error) {
            p, err := d.k8sWatcher.GetCachedPod(ns, podName)
            if err != nil {
                return "", err
            }
            return p.Annotations[bandwidth.EgressBandwidth], nil
        })
        ep.UpdateNoTrackRules(func(ns, podName string) (noTrackPort string, err error) {
            p, err := d.k8sWatcher.GetCachedPod(ns, podName)
            if err != nil {
                return "", err
            }
            return p.Annotations[annotation.NoTrack], nil
        })
    }

    regenTriggered := ep.UpdateLabels(ctx, addLabels, infoLabels, true)

    select {
    case <-ctx.Done():
        return d.errorDuringCreation(ep, fmt.Errorf("request cancelled while resolving identity"))
    default:
    }

    if !regenTriggered {
        regenMetadata := &regeneration.ExternalRegenerationMetadata{
            Reason:            "Initial build on endpoint creation",
            ParentContext:     ctx,
            RegenerationLevel: regeneration.RegenerateWithDatapathRewrite,
        }
        build, err := ep.SetRegenerateStateIfAlive(regenMetadata)
        if err != nil {
            return d.errorDuringCreation(ep, err)
        }
        if build {
            ep.Regenerate(regenMetadata)
        }
    }

    if epTemplate.SyncBuildEndpoint {
        if err := ep.WaitForFirstRegeneration(ctx); err != nil {
            return d.errorDuringCreation(ep, err)
        }
    }

    // The endpoint has been successfully created, stop the expiration
    // timers of all attached IPs
    if addressing := epTemplate.Addressing; addressing != nil {
        if uuid := addressing.IPV4ExpirationUUID; uuid != "" {
            if ip := net.ParseIP(addressing.IPV4); ip != nil {
                if err := d.ipam.StopExpirationTimer(ip, uuid); err != nil {
                    return d.errorDuringCreation(ep, err)
                }
            }
        }
    }

    return ep, 0, nil
}
