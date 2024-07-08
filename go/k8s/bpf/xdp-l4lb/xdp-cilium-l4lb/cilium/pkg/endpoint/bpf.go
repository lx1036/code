package endpoint

import (
	"fmt"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/endpoint/regeneration"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maps/eppolicymap"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maps/lxcmap"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maps/policymap"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"
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
	datapathRegenCtxt := regenContext.datapathRegenerationContext

	headerfileChanged, err = e.runPreCompilationSteps(regenContext)

	compilationExecuted, err = e.realizeBPFState(regenContext)
	if err != nil {
		return datapathRegenCtxt.epInfoCache.revision, compilationExecuted, err
	}

	if !datapathRegenCtxt.epInfoCache.IsHost() || option.Config.EnableHostFirewall {
		// Hook the endpoint into the endpoint and endpoint to policy tables then expose it
		stats.mapSync.Start()
		epErr := eppolicymap.WriteEndpoint(datapathRegenCtxt.epInfoCache, e.policyMap)
		err = lxcmap.WriteEndpoint(datapathRegenCtxt.epInfoCache)
		stats.mapSync.End(err == nil)
		if epErr != nil {
			e.logStatusLocked(BPF, Warning, fmt.Sprintf("Unable to sync EpToPolicy Map continue with Sockmap support: %s", epErr))
		}
		if err != nil {
			return 0, compilationExecuted, fmt.Errorf("Exposing new BPF failed: %s", err)
		}
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

// runPreCompilationSteps runs all of the regeneration steps that are necessary
// right before compiling the BPF for the given endpoint.
// The endpoint mutex must not be held.
//
// Returns whether the headerfile changed and/or an error.
func (e *Endpoint) runPreCompilationSteps(regenContext *regenerationContext) (headerfileChanged bool, preCompilationError error) {
	stats := &regenContext.Stats
	datapathRegenCtxt := regenContext.datapathRegenerationContext

	// policy map
	if e.policyMap == nil {
		e.policyMap, _, err = policymap.OpenOrCreate(e.policyMapPath())
		if err != nil {
			return false, err
		}

		// Synchronize the in-memory realized state with BPF map entries,
		// so that any potential discrepancy between desired and realized
		// state would be dealt with by the following e.syncPolicyMap.
		e.realizedPolicy.PolicyMapState, err = e.dumpPolicyMapToMapState()
		if err != nil {
			return false, err
		}
		//e.initPolicyMapPressureMetric()
		//e.updatePolicyMapPressureMetric()
	}

}
