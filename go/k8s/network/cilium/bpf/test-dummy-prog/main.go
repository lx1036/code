
//go:build linux
// +build linux


package main


import (
	"fmt"
	"golang.org/x/sys/unix"
	"k8s.io/klog/v2"
	"os"
	"runtime"
	"unsafe"
)

// ProgType is an enumeration for valid BPF program types
type ProgType int

// This enumeration must be in sync with enum bpf_prog_type in <linux/bpf.h>
const (
	ProgTypeUnspec ProgType = iota
	ProgTypeSocketFilter
	ProgTypeKprobe
	ProgTypeSchedCls
	ProgTypeSchedAct
	ProgTypeTracepoint
	ProgTypeXdp
	ProgTypePerfEvent
	ProgTypeCgroupSkb
	ProgTypeCgroupSock
	ProgTypeLwtIn
	ProgTypeLwtOut
	ProgTypeLwtXmit
	ProgTypeSockOps
	ProgTypeSkSkb
	ProgTypeCgroupDevice
	ProgTypeSkMsg
	ProgTypeRawTracepoint
	ProgTypeCgroupSockAddr
	ProgTypeLwtSeg6Local
	ProgTypeLircMode2
	ProgTypeSkReusePort
)

func (t ProgType) String() string {
	switch t {
	case ProgTypeSocketFilter:
		return "Socket filter"
	case ProgTypeKprobe:
		return "Kprobe"
	case ProgTypeSchedCls:
		return "Sched CLS"
	case ProgTypeSchedAct:
		return "Sched ACT"
	case ProgTypeTracepoint:
		return "Tracepoint"
	case ProgTypeXdp:
		return "XDP"
	case ProgTypePerfEvent:
		return "Perf event"
	case ProgTypeCgroupSkb:
		return "Cgroup skb"
	case ProgTypeCgroupSock:
		return "Cgroup sock"
	case ProgTypeLwtIn:
		return "LWT in"
	case ProgTypeLwtOut:
		return "LWT out"
	case ProgTypeLwtXmit:
		return "LWT xmit"
	case ProgTypeSockOps:
		return "Sock ops"
	case ProgTypeSkSkb:
		return "Socket skb"
	case ProgTypeCgroupDevice:
		return "Cgroup device"
	case ProgTypeSkMsg:
		return "Socket msg"
	case ProgTypeRawTracepoint:
		return "Raw tracepoint"
	case ProgTypeCgroupSockAddr:
		return "Cgroup sockaddr"
	case ProgTypeLwtSeg6Local:
		return "LWT seg6local"
	case ProgTypeLircMode2:
		return "LIRC"
	case ProgTypeSkReusePort:
		return "Socket SO_REUSEPORT"
	}
	
	return "Unknown"
}

const (
	// BPF syscall command constants. Must match enum bpf_cmd from linux/bpf.h
	BPF_MAP_CREATE          = 0
	BPF_MAP_LOOKUP_ELEM     = 1
	BPF_MAP_UPDATE_ELEM     = 2
	BPF_MAP_DELETE_ELEM     = 3
	BPF_MAP_GET_NEXT_KEY    = 4
	BPF_PROG_LOAD           = 5
	BPF_OBJ_PIN             = 6
	BPF_OBJ_GET             = 7
	BPF_PROG_ATTACH         = 8
	BPF_PROG_DETACH         = 9
	BPF_PROG_TEST_RUN       = 10
	BPF_PROG_GET_NEXT_ID    = 11
	BPF_MAP_GET_NEXT_ID     = 12
	BPF_PROG_GET_FD_BY_ID   = 13
	BPF_MAP_GET_FD_BY_ID    = 14
	BPF_OBJ_GET_INFO_BY_FD  = 15
	BPF_PROG_QUERY          = 16
	BPF_RAW_TRACEPOINT_OPEN = 17
	BPF_BTF_LOAD            = 18
	BPF_BTF_GET_FD_BY_ID    = 19
	BPF_TASK_FD_QUERY       = 20
	
	// BPF syscall attach types
	BPF_CGROUP_INET_INGRESS      = 0
	BPF_CGROUP_INET_EGRESS       = 1
	BPF_CGROUP_INET_SOCK_CREATE  = 2
	BPF_CGROUP_SOCK_OPS          = 3
	BPF_SK_SKB_STREAM_PARSER     = 4
	BPF_SK_SKB_STREAM_VERDICT    = 5
	BPF_CGROUP_DEVICE            = 6
	BPF_SK_MSG_VERDICT           = 7
	BPF_CGROUP_INET4_BIND        = 8
	BPF_CGROUP_INET6_BIND        = 9
	BPF_CGROUP_INET4_CONNECT     = 10
	BPF_CGROUP_INET6_CONNECT     = 11
	BPF_CGROUP_INET4_POST_BIND   = 12
	BPF_CGROUP_INET6_POST_BIND   = 13
	BPF_CGROUP_UDP4_SENDMSG      = 14
	BPF_CGROUP_UDP6_SENDMSG      = 15
	BPF_LIRC_MODE2               = 16
	BPF_FLOW_DISSECTOR           = 17
	BPF_CGROUP_SYSCTL            = 18
	BPF_CGROUP_UDP4_RECVMSG      = 19
	BPF_CGROUP_UDP6_RECVMSG      = 20
	BPF_CGROUP_INET4_GETPEERNAME = 29
	BPF_CGROUP_INET6_GETPEERNAME = 30
	BPF_CGROUP_INET4_GETSOCKNAME = 31
	BPF_CGROUP_INET6_GETSOCKNAME = 32
	
	// Flags for BPF_MAP_UPDATE_ELEM. Must match values from linux/bpf.h
	BPF_ANY     = 0
	BPF_NOEXIST = 1
	BPF_EXIST   = 2
	
	// Flags for BPF_MAP_CREATE. Must match values from linux/bpf.h
	BPF_F_NO_PREALLOC   = 1 << 0
	BPF_F_NO_COMMON_LRU = 1 << 1
	BPF_F_NUMA_NODE     = 1 << 2
	
	// Flags for BPF_PROG_QUERY
	BPF_F_QUERY_EFFECTVE = 1 << 0
	
	// Flags for accessing BPF object
	BPF_F_RDONLY = 1 << 3
	BPF_F_WRONLY = 1 << 4
	
	// Flag for stack_map, store build_id+offset instead of pointer
	BPF_F_STACK_BUILD_ID = 1 << 5
)

// @see https://github.com/cilium/cilium/blob/v1.11.1/daemon/cmd/kube_proxy_replacement.go#L456-L478
// @see https://github.com/cilium/cilium/blob/v1.11.1/pkg/bpf/bpf_linux.go#L611-L671
// go run .
// 测试 linux 内核是否支持 attach ebpf to cgroup
func main() {
	err := ProbeCgroupSupportTCP()
	if err != nil {
		klog.Fatalf(fmt.Sprintf("%+v", err))
	}
}

type bpfAttrProg struct {
	ProgType    uint32
	InsnCnt     uint32
	Insns       uintptr
	License     uintptr
	LogLevel    uint32
	LogSize     uint32
	LogBuf      uintptr
	KernVersion uint32
	Flags       uint32
	Name        [16]byte
	Ifindex     uint32
	AttachType  uint32
}

type bpfAttachProg struct {
	TargetFd    uint32
	AttachFd    uint32
	AttachType  uint32
	AttachFlags uint32
}

func ProbeCgroupSupportTCP() error {
	return TestDummyProg(ProgTypeCgroupSockAddr, BPF_CGROUP_INET4_CONNECT)
}

// TestDummyProg loads a minimal BPF program into the kernel and probes
// whether it succeeds in doing so. This can be used to bail out early
// in the daemon when a given type is not supported.
func TestDummyProg(progType ProgType, attachType uint32) error {
	var oldLim unix.Rlimit
	insns := []byte{
		// R0 = 1; EXIT
		0xb7, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
		0x95, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	license := []byte{'A', 'S', 'L', '2', '\x00'}
	bpfAttr := bpfAttrProg{
		ProgType:   uint32(progType),
		AttachType: uint32(attachType),
		InsnCnt:    uint32(len(insns) / 8),
		Insns:      uintptr(unsafe.Pointer(&insns[0])),
		License:    uintptr(unsafe.Pointer(&license[0])),
	}
	tmpLim := unix.Rlimit{
		Cur: unix.RLIM_INFINITY,
		Max: unix.RLIM_INFINITY,
	}
	err := unix.Getrlimit(unix.RLIMIT_MEMLOCK, &oldLim)
	if err != nil {
		return err
	}
	err = unix.Setrlimit(unix.RLIMIT_MEMLOCK, &tmpLim)
	if err != nil {
		return err
	}
	fd, _, errno := unix.Syscall(unix.SYS_BPF, BPF_PROG_LOAD,
		uintptr(unsafe.Pointer(&bpfAttr)),
		unsafe.Sizeof(bpfAttr))
	unix.Setrlimit(unix.RLIMIT_MEMLOCK, &oldLim)
	if errno == 0 {
		defer unix.Close(int(fd))
		bpfAttr := bpfAttachProg{
			TargetFd:   uint32(os.Stdin.Fd()),
			AttachFd:   uint32(fd),
			AttachType: attachType,
		}
		// We also need to go and probe the kernel whether we can actually
		// attach something to make sure CONFIG_CGROUP_BPF is compiled in.
		// The behavior is that when compiled in, we'll get a EBADF via
		// cgroup_bpf_prog_attach() -> cgroup_get_from_fd(), otherwise when
		// compiled out, we'll get EINVAL.
		ret, _, errno := unix.Syscall(unix.SYS_BPF, BPF_PROG_ATTACH,
			uintptr(unsafe.Pointer(&bpfAttr)),
			unsafe.Sizeof(bpfAttr))
		
		if int(ret) < 0 && errno != unix.EBADF {
			klog.Infof(fmt.Sprintf("err: %s", errno.Error())) // err: invalid argument
			
			return errno
		}
		return nil
	}
	
	klog.Infof(fmt.Sprintf("err: %s", errno.Error()))
	
	runtime.KeepAlive(&insns)
	runtime.KeepAlive(&license)
	runtime.KeepAlive(&bpfAttr)
	
	return errno
}
