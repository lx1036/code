package api

// AgentNotification specifies the type of agent notification
type AgentNotification uint32

const (
	AgentNotifyUnspec AgentNotification = iota
	AgentNotifyGeneric
	AgentNotifyStart
	AgentNotifyEndpointRegenerateSuccess
	AgentNotifyEndpointRegenerateFail
	AgentNotifyPolicyUpdated
	AgentNotifyPolicyDeleted
	AgentNotifyEndpointCreated
	AgentNotifyEndpointDeleted
	AgentNotifyIPCacheUpserted
	AgentNotifyIPCacheDeleted
	AgentNotifyServiceUpserted
	AgentNotifyServiceDeleted
)

// AgentNotifyMessage is a notification from the agent. It is similar to
// AgentNotify, but the notification is an unencoded struct. See the *Message
// constructors in this package for possible values.
type AgentNotifyMessage struct {
	Type         AgentNotification
	Notification interface{}
}

// Must be synchronized with <bpf/lib/common.h>
const (
	// 0-128 are reserved for BPF datapath events
	MessageTypeUnspec = iota

	// MessageTypeDrop is a BPF datapath notification carrying a DropNotify
	// which corresponds to drop_notify defined in bpf/lib/drop.h
	MessageTypeDrop

	// MessageTypeDebug is a BPF datapath notification carrying a DebugMsg
	// which corresponds to debug_msg defined in bpf/lib/dbg.h
	MessageTypeDebug

	// MessageTypeCapture is a BPF datapath notification carrying a DebugCapture
	// which corresponds to debug_capture_msg defined in bpf/lib/dbg.h
	MessageTypeCapture

	// MessageTypeTrace is a BPF datapath notification carrying a TraceNotify
	// which corresponds to trace_notify defined in bpf/lib/trace.h
	MessageTypeTrace

	// MessageTypePolicyVerdict is a BPF datapath notification carrying a PolicyVerdictNotify
	// which corresponds to policy_verdict_notify defined in bpf/lib/policy_log.h
	MessageTypePolicyVerdict

	// MessageTypeRecCapture is a BPF datapath notification carrying a RecorderCapture
	// which corresponds to capture_msg defined in bpf/lib/pcap.h
	MessageTypeRecCapture

	// 129-255 are reserved for agent level events

	// MessageTypeAccessLog contains a pkg/proxy/accesslog.LogRecord
	MessageTypeAccessLog = 129

	// MessageTypeAgent is an agent notification carrying a AgentNotify
	MessageTypeAgent = 130
)
