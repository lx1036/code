package format

import (
    "bytes"
    "encoding/binary"
    "fmt"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/byteorder"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/hubble/parser/getters"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/monitor"
    monitorAPI "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/monitor/api"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/monitor/payload"
)

// Verbosity levels for formatting output.
type Verbosity uint8

const (
    msgSeparator = "------------------------------------------------------------------------------"

    // INFO is the level of verbosity in which summaries of Drop and Capture
    // messages are printed out when the monitor is invoked
    INFO Verbosity = iota + 1
    // DEBUG is the level of verbosity in which more information about packets
    // is printed than in INFO mode. Debug, Drop, and Capture messages are printed.
    DEBUG
    // VERBOSE is the level of verbosity in which the most information possible
    // about packets is printed out. Currently is not utilized.
    VERBOSE
    // JSON is the level of verbosity in which event information is printed out in json format
    JSON
)

// MonitorFormatter filters and formats monitor messages from a buffer.
type MonitorFormatter struct {
    EventTypes monitorAPI.MessageTypeFilter
    FromSource Uint16Flags
    ToDst      Uint16Flags
    Related    Uint16Flags
    Hex        bool
    JSONOutput bool
    Verbosity  Verbosity
    Numeric    bool

    linkMonitor getters.LinkGetter
}

// FormatEvent formats an event from the specified payload to stdout.
//
// Returns true if the event was successfully printed, false otherwise.
func (m *MonitorFormatter) FormatEvent(pl *payload.Payload) bool {
    switch pl.Type {
    case payload.EventSample:
        m.FormatSample(pl.Data, pl.CPU)
    case payload.RecordLost:
        LostEvent(pl.Lost, pl.CPU)
    default:
        return false
    }

    return true
}

// FormatSample prints an event from the provided raw data slice to stdout.
//
// For most monitor event types, 'data' corresponds to the 'data' field in
// bpf.PerfEventSample. Exceptions are MessageTypeAccessLog and
// MessageTypeAgent.
func (m *MonitorFormatter) FormatSample(data []byte, cpu int) {
    prefix := fmt.Sprintf("CPU %02d:", cpu)
    messageType := data[0]

    switch messageType {
    case monitorAPI.MessageTypeDrop:
        //m.dropEvents(prefix, data)
    case monitorAPI.MessageTypeDebug:
        m.debugEvents(prefix, data)
    case monitorAPI.MessageTypeCapture:
        //m.captureEvents(prefix, data)
    case monitorAPI.MessageTypeTrace:
        //m.traceEvents(prefix, data)
    case monitorAPI.MessageTypeAccessLog:
        //m.logRecordEvents(prefix, data)
    case monitorAPI.MessageTypeAgent:
        //m.agentEvents(prefix, data)
    case monitorAPI.MessageTypePolicyVerdict:
        //m.policyVerdictEvents(prefix, data)
    case monitorAPI.MessageTypeRecCapture:
        //m.recorderCaptureEvents(prefix, data)
    default:
        fmt.Printf("%s Unknown event: %+v\n", prefix, data)
    }
}

// debugEvents prints out all the debug messages.
func (m *MonitorFormatter) debugEvents(prefix string, data []byte) {
    dm := monitor.DebugMsg{}

    if err := binary.Read(bytes.NewReader(data), byteorder.Native, &dm); err != nil {
        fmt.Printf("Error while parsing debug message: %s\n", err)
    }
    if m.match(monitorAPI.MessageTypeDebug, dm.Source, 0) {
        switch m.Verbosity {
        case INFO:
            dm.DumpInfo(data)
        case JSON:
            dm.DumpJSON(prefix, m.linkMonitor)
        default:
            dm.Dump(prefix, m.linkMonitor)
        }
    }
}

func NewMonitorFormatter(verbosity Verbosity, linkMonitor getters.LinkGetter) *MonitorFormatter {
    return &MonitorFormatter{
        Hex:         false,
        EventTypes:  monitorAPI.MessageTypeFilter{},
        FromSource:  Uint16Flags{},
        ToDst:       Uint16Flags{},
        Related:     Uint16Flags{},
        JSONOutput:  false,
        Verbosity:   verbosity,
        Numeric:     bool(monitor.DisplayLabel),
        linkMonitor: linkMonitor,
    }
}

// LostEvent formats a lost event using the specified payload parameters.
func LostEvent(lost uint64, cpu int) {
    fmt.Printf("CPU %02d: Lost %d events\n", cpu, lost)
}
