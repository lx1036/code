package endpoint

import (
    "context"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/completion"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/revert"
)

// RegenerationContext provides context to regenerate() calls to determine
// the caller, and which specific aspects to regeneration are necessary to
// update the datapath to implement the new behavior.
type regenerationContext struct {
    // Reason provides context to source for the regeneration, which is
    // used to generate useful log messages.
    Reason string

    // Stats are collected during the endpoint regeneration and provided
    // back to the caller
    Stats regenerationStatistics

    // DoneFunc must be called when the most resource intensive portion of
    // the regeneration is done
    DoneFunc func()

    datapathRegenerationContext *datapathRegenerationContext

    parentContext context.Context

    cancelFunc context.CancelFunc
}

// datapathRegenerationContext contains information related to regenerating the
// datapath (BPF, proxy, etc.).
type datapathRegenerationContext struct {
    bpfHeaderfilesHash string
    epInfoCache        *epInfoCache
    proxyWaitGroup     *completion.WaitGroup
    ctCleaned          chan struct{}
    completionCtx      context.Context
    completionCancel   context.CancelFunc
    currentDir         string
    nextDir            string
    regenerationLevel  regeneration.DatapathRegenerationLevel

    finalizeList revert.FinalizeList
    revertStack  revert.RevertStack
}
