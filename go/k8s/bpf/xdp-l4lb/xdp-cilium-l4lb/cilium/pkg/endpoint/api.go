package endpoint

import (
    "context"

    "github.com/cilium/cilium/pkg/addressing"
    "github.com/cilium/cilium/pkg/identity/cache"
    "github.com/cilium/cilium/pkg/labelsfilter"
    "sort"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/api/v1/models"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/endpoint/regeneration"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/labels"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/labels/model"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/mac"
)

// GetModel returns the API model of endpoint e.
func (e *Endpoint) GetModel() *models.Endpoint {
    if e == nil {
        return nil
    }
    // NOTE: Using rlock on mutex directly because GetModelRLocked handles removed endpoint properly
    e.mutex.RLock()
    defer e.mutex.RUnlock()

    return e.GetModelRLocked()
}

func (e *Endpoint) GetModelRLocked() *models.Endpoint {
    if e == nil {
        return nil
    }

    // `cilium endpoint log 3994`
    statusLog := e.status.GetModel()
    if len(statusLog) > 0 {
        statusLog = statusLog[:1]
    }

    lblMdl := model.NewModel(&e.OpLabels)
    // Sort these slices since they come out in random orders. This allows
    // reflect.DeepEqual to succeed.
    sort.StringSlice(lblMdl.Realized.User).Sort()
    sort.StringSlice(lblMdl.Disabled).Sort()
    sort.StringSlice(lblMdl.SecurityRelevant).Sort()
    sort.StringSlice(lblMdl.Derived).Sort()

    controllerMdl := e.controllers.GetStatusModel()
    sort.Slice(controllerMdl, func(i, j int) bool { return controllerMdl[i].Name < controllerMdl[j].Name })

    spec := &models.EndpointConfigurationSpec{
        LabelConfiguration: lblMdl.Realized,
    }

    if e.Options != nil {
        spec.Options = *e.Options.GetMutableModel()
    }

    return &models.Endpoint{
        ID:   int64(e.ID),
        Spec: spec,
        Status: &models.EndpointStatus{
            // FIXME GH-3280 When we begin implementing revision numbers this will
            // diverge from models.Endpoint.Spec to reflect the in-datapath config
            Realized:            spec,
            Identity:            identitymodel.CreateModel(e.SecurityIdentity),
            Labels:              lblMdl,
            Networking:          e.getModelNetworkingRLocked(),
            ExternalIdentifiers: e.getModelEndpointIdentitiersRLocked(),
            // FIXME GH-3280 When we begin returning endpoint revisions this should
            // change to return the configured and in-datapath policies.
            Policy:      e.GetPolicyModel(),
            Log:         statusLog,
            Controllers: controllerMdl,
            State:       e.getModelCurrentStateRLocked(), // TODO: Validate
            Health:      e.getHealthModel(),
            NamedPorts:  e.getNamedPortsModel(),
        },
    }
}

// NewEndpointFromChangeModel creates a new endpoint from a request
func NewEndpointFromChangeModel(ctx context.Context, owner regeneration.Owner, policyGetter policyRepoGetter, namedPortsGetter namedPortsGetter, proxy EndpointProxy, allocator cache.IdentityAllocator, base *models.EndpointChangeRequest) (*Endpoint, error) {
    if base == nil {
        return nil, nil
    }

    ep := createEndpoint(owner, policyGetter, namedPortsGetter, proxy, allocator, uint16(base.ID), base.InterfaceName)
    ep.ifIndex = int(base.InterfaceIndex)
    ep.containerName = base.ContainerName
    ep.containerID = base.ContainerID
    ep.dockerNetworkID = base.DockerNetworkID
    ep.dockerEndpointID = base.DockerEndpointID
    ep.K8sPodName = base.K8sPodName
    ep.K8sNamespace = base.K8sNamespace

    if base.Mac != "" {
        m, err := mac.ParseMAC(base.Mac)
        if err != nil {
            return nil, err
        }
        ep.mac = m
    }

    if base.HostMac != "" {
        m, err := mac.ParseMAC(base.HostMac)
        if err != nil {
            return nil, err
        }
        ep.nodeMAC = m
    }

    if base.Addressing != nil {
        if ip := base.Addressing.IPV6; ip != "" {
            ip6, err := addressing.NewCiliumIPv6(ip)
            if err != nil {
                return nil, err
            }
            ep.IPv6 = ip6
        }

        if ip := base.Addressing.IPV4; ip != "" {
            ip4, err := addressing.NewCiliumIPv4(ip)
            if err != nil {
                return nil, err
            }
            ep.IPv4 = ip4
        }
    }

    if base.DatapathConfiguration != nil {
        ep.DatapathConfiguration = *base.DatapathConfiguration
    }

    if base.Labels != nil {
        lbls := labels.NewLabelsFromModel(base.Labels)
        identityLabels, infoLabels := labelsfilter.Filter(lbls)
        ep.OpLabels.OrchestrationIdentity = identityLabels
        ep.OpLabels.OrchestrationInfo = infoLabels
    }

    ep.setState(State(base.State), "Endpoint creation")

    return ep, nil
}
