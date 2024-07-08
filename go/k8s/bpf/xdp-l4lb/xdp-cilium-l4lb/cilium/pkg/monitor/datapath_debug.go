package monitor

import (
	"fmt"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/hubble/parser/getters"
)

/*
// must be in sync with <bpf/lib/dbg.h>
struct debug_msg {
    __u8		type;
	__u8		subtype;
	__u16		source;
	__u32		hash;
	__u32		arg1;
	__u32		arg2;
	__u32		arg3;
};
*/

// must be in sync with <bpf/lib/dbg.h>
const (
	DbgUnspec = iota
	DbgGeneric
	DbgLocalDelivery
	DbgEncap
	DbgLxcFound
	DbgPolicyDenied
	DbgCtLookup
	DbgCtLookupRev
	DbgCtMatch
	DbgCtCreated
	DbgCtCreated2
	DbgIcmp6Handle
	DbgIcmp6Request
	DbgIcmp6Ns
	DbgIcmp6TimeExceeded
	DbgCtVerdict
	DbgDecap
	DbgPortMap
	DbgErrorRet
	DbgToHost
	DbgToStack
	DbgPktHash
	DbgLb6LookupFrontend
	DbgLb6LookupFrontendFail
	DbgLb6LookupBackendSlot
	DbgLb6LookupBackendSlotSuccess
	DbgLb6LookupBackendSlotV2Fail
	DbgLb6LookupBackendFail
	DbgLb6ReverseNatLookup
	DbgLb6ReverseNat
	DbgLb4LookupFrontend
	DbgLb4LookupFrontendFail
	DbgLb4LookupBackendSlot
	DbgLb4LookupBackendSlotSuccess
	DbgLb4LookupBackendSlotV2Fail
	DbgLb4LookupBackendFail
	DbgLb4ReverseNatLookup
	DbgLb4ReverseNat
	DbgLb4LoopbackSnat
	DbgLb4LoopbackSnatRev
	DbgCtLookup4
	DbgRRBackendSlotSel
	DbgRevProxyLookup
	DbgRevProxyFound
	DbgRevProxyUpdate
	DbgL4Policy
	DbgNetdevInCluster
	DbgNetdevEncap4
	DbgCTLookup41
	DbgCTLookup42
	DbgCTCreated4
	DbgCTLookup61
	DbgCTLookup62
	DbgCTCreated6
	DbgSkipProxy
	DbgL4Create
	DbgIPIDMapFailed4
	DbgIPIDMapFailed6
	DbgIPIDMapSucceed4
	DbgIPIDMapSucceed6
	DbgLbStaleCT
	DbgInheritIdentity
	DbgSkLookup4
	DbgSkLookup6
	DbgSkAssign
	DbgL7LB
)

// DebugMsg is the message format of the debug message found in the BPF ring buffer
type DebugMsg struct {
	Type    uint8
	SubType uint8
	Source  uint16
	Hash    uint32
	Arg1    uint32
	Arg2    uint32
	Arg3    uint32
}

// DumpInfo prints a summary of a subset of the debug messages which are related
// to sending, not processing, of packets.
func (n *DebugMsg) DumpInfo(data []byte) {
}

// Dump prints the debug message in a human readable format.
func (n *DebugMsg) Dump(prefix string, linkMonitor getters.LinkGetter) {
	fmt.Printf("%s MARK %#x FROM %d DEBUG: %s\n", prefix, n.Hash, n.Source, n.Message(linkMonitor))
}

func (n *DebugMsg) getJSON(cpuPrefix string, linkMonitor getters.LinkGetter) string {
	return fmt.Sprintf(`{"cpu":%q,"type":"debug","message":%q}`,
		cpuPrefix, n.Message(linkMonitor))
}

// DumpJSON prints notification in json format
func (n *DebugMsg) DumpJSON(cpuPrefix string, linkMonitor getters.LinkGetter) {
	fmt.Println(n.getJSON(cpuPrefix, linkMonitor))
}

// Message returns the debug message in a human-readable format
func (n *DebugMsg) Message(linkMonitor getters.LinkGetter) string {
	switch n.SubType {
	case DbgGeneric:
		return fmt.Sprintf("No message, arg1=%d (%#x) arg2=%d (%#x)", n.Arg1, n.Arg1, n.Arg2, n.Arg2)
	default:
		return fmt.Sprintf("Unknown message type=%d arg1=%d arg2=%d", n.SubType, n.Arg1, n.Arg2)
	}
}
