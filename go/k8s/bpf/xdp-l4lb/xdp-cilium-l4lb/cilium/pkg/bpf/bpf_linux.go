package bpf

import (
	"fmt"
	"github.com/cilium/cilium/pkg/metrics"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"runtime"
	"unsafe"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/spanstat"
)

// This struct must be in sync with union bpf_attr's anonymous struct used by
// BPF_OBJ_*_ commands
type bpfAttrObjOp struct {
	pathname uint64
	fd       uint32
	pad0     [4]byte
}

// This struct must be in sync with union bpf_attr's anonymous struct used by
// BPF_MAP_*_ELEM commands
type bpfAttrMapOpElem struct {
	mapFd uint32
	pad0  [4]byte
	key   uint64
	value uint64 // union: value or next_key
	flags uint64
}

// ObjGet map path -> fd
func ObjGet(pathname string) (int, error) {
	pathStr, err := unix.BytePtrFromString(pathname)
	if err != nil {
		return 0, fmt.Errorf("Unable to convert pathname %q to byte pointer: %w", pathname, err)
	}
	uba := bpfAttrObjOp{
		pathname: uint64(uintptr(unsafe.Pointer(pathStr))),
	}

	var duration *spanstat.SpanStat
	if option.Config.MetricsConfig.BPFSyscallDurationEnabled {
		duration = spanstat.Start()
	}
	fd, _, errno := unix.Syscall(
		unix.SYS_BPF,
		BPF_OBJ_GET,
		uintptr(unsafe.Pointer(&uba)),
		unsafe.Sizeof(uba),
	)
	runtime.KeepAlive(pathStr)
	runtime.KeepAlive(&uba)
	if option.Config.MetricsConfig.BPFSyscallDurationEnabled {
		metrics.BPFSyscallDuration.WithLabelValues(metricOpObjGet,
			metrics.Errno2Outcome(errno)).Observe(duration.End(errno == 0).Total().Seconds())
	}

	if fd == 0 || errno != 0 {
		return 0, &os.PathError{
			Op:   "Unable to get object",
			Err:  errno,
			Path: pathname,
		}
	}

	return int(fd), nil
}

// GetFirstKey fetches the first key in the map. If there are no keys in the
// map, io.EOF is returned.
func GetFirstKey(fd int, nextKey unsafe.Pointer) error {
	uba := bpfAttrMapOpElem{
		mapFd: uint32(fd),
		key:   0, // NULL -> Get first element
		value: uint64(uintptr(nextKey)),
	}

	ret := GetNextKeyFromPointers(fd, unsafe.Pointer(&uba), unsafe.Sizeof(uba))
	runtime.KeepAlive(nextKey)
	return ret
}

// GetNextKeyFromPointers stores, in nextKey, the next key after the key of the
// map in fd. When there are no more keys, io.EOF is returned.
func GetNextKeyFromPointers(fd int, structPtr unsafe.Pointer, sizeOfStruct uintptr) error {
	var duration *spanstat.SpanStat
	if option.Config.MetricsConfig.BPFSyscallDurationEnabled {
		duration = spanstat.Start()
	}
	ret, _, err := unix.Syscall(
		unix.SYS_BPF,
		BPF_MAP_GET_NEXT_KEY,
		uintptr(structPtr),
		sizeOfStruct,
	)
	runtime.KeepAlive(structPtr)
	if option.Config.MetricsConfig.BPFSyscallDurationEnabled {
		metrics.BPFSyscallDuration.WithLabelValues(metricOpGetNextKey,
			metrics.Errno2Outcome(err)).Observe(duration.End(err == 0).Total().Seconds())
	}

	// BPF_MAP_GET_NEXT_KEY returns ENOENT when all keys have been iterated
	// translate that to io.EOF to signify there are no next keys
	if err == unix.ENOENT {
		return io.EOF
	}

	if ret != 0 || err != 0 {
		return fmt.Errorf("Unable to get next key from map with file descriptor %d: %s", fd, err)
	}

	return nil
}

// LookupElementFromPointers looks up for the map value stored in fd with the given key. The value
// is stored in the value unsafe.Pointer.
func LookupElementFromPointers(fd int, structPtr unsafe.Pointer, sizeOfStruct uintptr) error {
	var duration *spanstat.SpanStat
	if option.Config.MetricsConfig.BPFSyscallDurationEnabled {
		duration = spanstat.Start()
	}
	ret, _, err := unix.Syscall(
		unix.SYS_BPF,
		BPF_MAP_LOOKUP_ELEM,
		uintptr(structPtr),
		sizeOfStruct,
	)
	runtime.KeepAlive(structPtr)
	if option.Config.MetricsConfig.BPFSyscallDurationEnabled {
		metrics.BPFSyscallDuration.WithLabelValues(metricOpLookup,
			metrics.Errno2Outcome(err)).Observe(duration.End(err == 0).Total().Seconds())
	}

	if ret != 0 || err != 0 {
		return fmt.Errorf("Unable to lookup element in map with file descriptor %d: %w", fd, err)
	}

	return nil
}
