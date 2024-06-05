package defaults

const (
    DefaultMapRoot   = "/sys/fs/bpf"
    DefaultMapPrefix = "tc/globals"

    // MonitorBufferPages is the default number of pages to use for the
    // ring buffer interacting with the kernel
    MonitorBufferPages = 64

    // RuntimePath is the default path to the runtime directory
    RuntimePath = "/var/run/cilium"

    // MonitorSockPath1_2 is the path to the UNIX domain socket used to
    // distribute BPF and agent events to listeners.
    // This is the 1.2 protocol version.
    MonitorSockPath1_2 = RuntimePath + "/monitor1_2.sock"
)
