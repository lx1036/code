package endpoint

import (
    "sort"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/api/v1/models"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/labels/model"
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
