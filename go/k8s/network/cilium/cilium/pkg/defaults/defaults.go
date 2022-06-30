package defaults

// Base
const (
	EnvNodeNameSpec = "K8S_NODE_NAME"
)

// BPF
const (
	// DefaultMapRoot is the default path where BPFFS should be mounted
	DefaultMapRoot = "/sys/fs/bpf"

	// DefaultMapPrefix is the default prefix for all BPF maps.
	DefaultMapPrefix = "tc/globals"

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
)
