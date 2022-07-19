package defaults

import "time"

// Base
const (
	EnvNodeNameSpec = "K8S_NODE_NAME"

	// ExecTimeout is a timeout for executing commands.
	ExecTimeout = 300 * time.Second

	// LibraryPath is the default path to the cilium libraries directory
	LibraryPath = "/var/lib/cilium"

	// HostDevice is the name of the device that connects the cilium IP
	// space with the host's networking model
	HostDevice = "cilium_host"
)

// BPF
const (
	// DefaultMapRoot is the default path where BPFFS should be mounted
	DefaultMapRoot = "/sys/fs/bpf"

	// DefaultMapPrefix is the default prefix for all BPF maps.
	DefaultMapPrefix = "tc/globals"

	// RuntimePathRights are the default access rights of the RuntimePath directory
	RuntimePathRights = 0775

	////////////////////////////// bpf //////////////////////////////
	// RestoreV4Addr is used as match for cilium_host v4 address
	RestoreV4Addr = "cilium.v4.internal.raw "

	// CHeaderFileName is the name of the C header file for BPF programs for a
	// particular endpoint.
	CHeaderFileName = "ep_config.h"
	// OldCHeaderFileName is the previous name of the C header file for BPF
	// programs for a particular endpoint. It can be removed once Cilium v1.8
	// is the oldest supported version.
	OldCHeaderFileName = "lxc_config.h"
	// HostObjFileName is the name of the host object file.
	HostObjFileName = "bpf_host.o"
	// CiliumCHeaderPrefix is the prefix using when printing/writing an endpoint in a
	// base64 form.
	CiliumCHeaderPrefix = "CILIUM_BASE64_"

	// TemplatesDir is the default path for the compiled template objects relative to StateDir
	TemplatesDir = "templates"

	// BpfDir is the default path for template files relative to LibDir
	BpfDir = "bpf"
)

// Cgroup
const (
	// DefaultCgroupRoot is the default path where cilium cgroup2 should be mounted
	DefaultCgroupRoot = "/var/run/cilium/cgroupv2"
)
