package endpoint

import (
    "github.com/cilium/cilium/pkg/option"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf/maps/policymap"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/endpoint/regeneration"
)

// policyMapPath returns the path to the policy map of endpoint.
func (e *Endpoint) policyMapPath() string {
    return bpf.LocalMapPath(policymap.MapName, e.ID)
}

// InitPolicyMap creates the policy map in the kernel.
func (e *Endpoint) InitPolicyMap() error {
    _, err := policymap.Create(e.policyMapPath())
    return err
}

// regenerateBPF rewrites all headers and updates all BPF maps to reflect the
// specified endpoint.
func (e *Endpoint) regenerateBPF(regenContext *regenerationContext) (revnum uint64, stateDirComplete bool, reterr error) {

    compilationExecuted, err = e.realizeBPFState(regenContext)
    if err != nil {
        return datapathRegenCtxt.epInfoCache.revision, compilationExecuted, err
    }

}

func (e *Endpoint) realizeBPFState(regenContext *regenerationContext) (compilationExecuted bool, err error) {
    stats := &regenContext.Stats
    datapathRegenCtxt := regenContext.datapathRegenerationContext

    e.getLogger().WithField(fieldRegenLevel, datapathRegenCtxt.regenerationLevel).Debug("Preparing to compile BPF")

    if datapathRegenCtxt.regenerationLevel > regeneration.RegenerateWithoutDatapath {
        // Compile and install BPF programs for this endpoint
        if e.Options.IsEnabled(option.Debug) {
            // TODO
        }

        if datapathRegenCtxt.regenerationLevel == regeneration.RegenerateWithDatapathRebuild {
            err = e.owner.Datapath().Loader().CompileAndLoad(datapathRegenCtxt.completionCtx, datapathRegenCtxt.epInfoCache, &stats.datapathRealization)
            e.getLogger().WithError(err).Info("Regenerated endpoint BPF program")
            compilationExecuted = true
        } else if datapathRegenCtxt.regenerationLevel == regeneration.RegenerateWithDatapathRewrite {
            err = e.owner.Datapath().Loader().CompileOrLoad(datapathRegenCtxt.completionCtx, datapathRegenCtxt.epInfoCache, &stats.datapathRealization)
            if err == nil {
                e.getLogger().Info("Rewrote endpoint BPF program")
            } else {
                e.getLogger().WithError(err).Error("Error while rewriting endpoint BPF program")
            }
            compilationExecuted = true
        } else { // RegenerateWithDatapathLoad
            err = e.owner.Datapath().Loader().ReloadDatapath(datapathRegenCtxt.completionCtx, datapathRegenCtxt.epInfoCache, &stats.datapathRealization)
            if err == nil {
                e.getLogger().Info("Reloaded endpoint BPF program")
            } else {
                e.getLogger().WithError(err).Error("Error while reloading endpoint BPF program")
            }
        }

        if err != nil {
            return compilationExecuted, err
        }
        e.bpfHeaderfileHash = datapathRegenCtxt.bpfHeaderfilesHash

    } else {
        e.getLogger().WithField(logfields.BPFHeaderfileHash, datapathRegenCtxt.bpfHeaderfilesHash).
            Debug("BPF header file unchanged, skipping BPF compilation and installation")
    }

    return compilationExecuted, nil
}

// writeHeaderfile writes the lxc_config.h header file of an endpoint.
//
// e.mutex must be write-locked.
func (e *Endpoint) writeHeaderfile(prefix string) error {

}
