package loader

import (
    "context"
    "fmt"
    "github.com/cilium/cilium/pkg/bpf"
    "strconv"
)

// replaceDatapath replaces the qdisc and BPF program for an endpoint or XDP program.
//
// When successful, returns a finalizer to allow the map cleanup operation to be
// deferred by the caller. On error, any maps pending migration are immediately
// re-pinned to their original paths and a finalizer is not returned.
//
// When replacing multiple programs from the same ELF in a loop, the finalizer
// should only be run when all the interface's programs have been replaced
// since they might share one or more tail call maps.
//
// For example, this is the case with from-netdev and to-netdev. If eth0:to-netdev
// gets its program and maps replaced and unpinned, its eth0:from-netdev counterpart
// will miss tail calls (and drop packets) until it has been replaced as well.
func replaceDatapath(ctx context.Context, ifName, objPath, progSec, progDirection string, xdp bool, xdpMode string) (func(), error) {
    var (
        loaderProg string
        args       []string
    )

    if !xdp {
        if err := replaceQdisc(ifName); err != nil {
            return nil, fmt.Errorf("Failed to replace Qdisc for %s: %s", ifName, err)
        }
    }

    // Temporarily rename bpffs pins of maps whose definitions have changed in
    // a new version of a datapath ELF.
    if err := bpf.StartBPFFSMigration(bpf.MapPrefixPath(), objPath); err != nil {
        return nil, fmt.Errorf("Failed to start bpffs map migration: %w", err)
    }

    // FIXME: replace exec with native call
    if xdp {
        loaderProg = "ip"
        args = []string{"-force", "link", "set", "dev", ifName, xdpMode,
            "obj", objPath, "sec", progSec}
    } else {
        loaderProg = "tc"

        tcPrio := strconv.Itoa(option.Config.TCFilterPriority)
        log.Debugf("tc filter using priority %s for interface %s", tcPrio, ifName)
        args = []string{"filter", "replace", "dev", ifName, progDirection,
            "prio", tcPrio, "handle", "1", "bpf", "da", "obj", objPath,
            "sec", progSec,
        }
    }

    // If the iproute2 call below is successful, any 'pending' map pins will be removed.
    // If not, any pending maps will be re-pinned back to their initial paths.
    cmd := exec.CommandContext(ctx, loaderProg, args...).WithFilters(libbpfFixupMsg)
    if _, err := cmd.CombinedOutput(log, true); err != nil {
        // Program/object replacement unsuccessful, revert bpffs migration.
        if err := bpf.FinalizeBPFFSMigration(bpf.MapPrefixPath(), objPath, true); err != nil {
            return nil, fmt.Errorf("Failed to revert bpffs map migration: %w", err)
        }
        return nil, fmt.Errorf("Failed to load prog with %s: %w", loaderProg, err)
    }

    finalize := func() {
        l := log.WithField("device", ifName).WithField("objPath", objPath)
        l.Debug("Finalizing bpffs map migration")
        if err := bpf.FinalizeBPFFSMigration(bpf.MapPrefixPath(), objPath, false); err != nil {
            l.WithError(err).Error("Could not finalize bpffs map migration")
        }
    }

    return finalize, nil
}
