package main

import (
    "encoding/binary"
    "fmt"
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
    "github.com/containernetworking/plugins/pkg/ns"
    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
    "golang.org/x/sys/unix"
    "io"
    "os"
    "path/filepath"
)

const (
    NetnsPath = "/proc/self/ns/net"
    BPFFsPath = "/sys/fs/bpf"
)

func init() {
    rootCmd.AddCommand(loadCmd)
    rootCmd.AddCommand(unloadCmd)

}

// CGO_ENABLED=0 go run . load
var loadCmd = &cobra.Command{
    Use:     "load",
    Example: "load",
    Run: func(cmd *cobra.Command, args []string) {
        dispatcher, err := CreateDispatcher()
        if err != nil {
            logrus.Errorf("err: %v", err)
            return
        }
        defer dispatcher.Close()
    },
}

var unloadCmd = &cobra.Command{
    Use:     "unload",
    Example: "unload",
    Run: func(cmd *cobra.Command, args []string) {
        UnloadDispatcher()
    },
}

type Dispatcher struct {
    bindings     *ebpf.Map
    destinations *Destinations
}

func CreateDispatcher() (*Dispatcher, error) {
    var err error
    closeOnError := func(c io.Closer) {
        if err != nil {
            c.Close()
        }
    }

    netNs, pinPath, err := openNetNS(NetnsPath, BPFFsPath)
    if err != nil {
        return nil, err
    }
    defer netNs.Close()

    err = os.MkdirAll(pinPath, 0750)
    if err != nil {
        return nil, err
    }

    objs := dispatcherObjects{}
    opts := &ebpf.CollectionOptions{
        Programs: ebpf.ProgramOptions{
            LogLevel: ebpf.LogLevelInstruction,
            LogSize:  64 * 1024 * 1024, // 64M
        },
        Maps: ebpf.MapOptions{
            PinPath: pinPath, // /sys/fs/bpf/xxx_dispatcher/
        },
    }
    spec, err := loadDispatcher()
    if err != nil {
        return nil, err
    }
    var specs dispatcherSpecs
    err = spec.Assign(&specs)
    if err != nil {
        return nil, err
    }
    // check max_entries of sockets/destinations/destination_metrics
    maxSockets := specs.Sockets.MaxEntries
    for _, m := range []*ebpf.MapSpec{
        specs.Destinations,
        specs.DestinationMetrics,
    } {
        if m.MaxEntries != maxSockets {
            return nil, fmt.Errorf("map %q has %d max entries instead of %d", m.Name, m.MaxEntries, maxSockets)
        }
    }
    // set key_size/value_size of destinations
    specs.Destinations.KeySize = uint32(binary.Size(destinationKey{}))
    specs.Destinations.ValueSize = uint32(binary.Size(destinationValue{}))
    err = spec.LoadAndAssign(&objs, opts)
    if err != nil {
        return nil, err
    }
    defer objs.dispatcherPrograms.Close()
    defer closeOnError(&objs.dispatcherMaps) // 只有 error 才 close bpf maps
    // 因为 pin program，所以可以 close dispatcher program
    if err = objs.dispatcherPrograms.Dispatcher.Pin(fmt.Sprintf("%s/program", pinPath)); err != nil {
        return nil, err
    }

    // Attach
    l, err := link.AttachNetNs(int(netNs.Fd()), objs.dispatcherPrograms.Dispatcher)
    if err != nil {
        return nil, fmt.Errorf("attach program to netns %s: %s", netNs.Path(), err)
    }
    defer l.Close()
    // 因为 pin link，所以可以 close dispatcher link
    if err = l.Pin(fmt.Sprintf("%s/link", pinPath)); err != nil {
        return nil, err
    }

    return &Dispatcher{
        bindings:     objs.dispatcherMaps.Bindings,
        destinations: NewDestinations(objs.dispatcherMaps),
    }, nil
}

func UnloadDispatcher() {
    netNs, pinPath, err := openNetNS(NetnsPath, BPFFsPath)
    if err != nil {
        logrus.Errorf("[UnloadDispatcher]err: %v", err)
        return
    }
    defer netNs.Close()

    if err = os.RemoveAll(pinPath); err != nil {
        logrus.Errorf("[UnloadDispatcher]err: %v", err)
        return
    }
}

// OpenDispatcher loads an existing dispatcher from a namespace.
func OpenDispatcher(readOnly bool) (*Dispatcher, error) {
    netNs, pinPath, err := openNetNS(NetnsPath, BPFFsPath)
    if err != nil {
        return nil, err
    }
    defer netNs.Close()

    objs := dispatcherObjects{}
    spec, err := loadDispatcher()
    if err != nil {
        return nil, err
    }
    var specs dispatcherSpecs
    err = spec.Assign(&specs)
    if err != nil {
        return nil, err
    }
    // check max_entries of sockets/destinations/destination_metrics
    maxSockets := specs.Sockets.MaxEntries
    for _, m := range []*ebpf.MapSpec{
        specs.Destinations,
        specs.DestinationMetrics,
    } {
        if m.MaxEntries != maxSockets {
            return nil, fmt.Errorf("map %q has %d max entries instead of %d", m.Name, m.MaxEntries, maxSockets)
        }
    }
    // set key_size/value_size of destinations
    specs.Destinations.KeySize = uint32(binary.Size(destinationKey{}))
    specs.Destinations.ValueSize = uint32(binary.Size(destinationValue{}))

    err = spec.LoadAndAssign(&objs, &ebpf.CollectionOptions{
        Maps: ebpf.MapOptions{
            PinPath: pinPath, // /sys/fs/bpf/xxx_dispatcher/
            LoadPinOptions: ebpf.LoadPinOptions{
                ReadOnly: readOnly, // 注意这个参数
            },
        },
    })
    if err != nil {
        return nil, err
    }

    return &Dispatcher{
        bindings:     objs.dispatcherMaps.Bindings,
        destinations: NewDestinations(objs.dispatcherMaps),
    }, nil
}

// Close frees bpf maps
//
// It does not remove the dispatcher, see UnloadDispatcher.
func (dispatcher *Dispatcher) Close() error {
    // No need to lock the state, since we don't modify it here.
    if err := dispatcher.bindings.Close(); err != nil {
        return fmt.Errorf("can't close BPF objects: %s", err)
    }
    if err := dispatcher.destinations.Close(); err != nil {
        return fmt.Errorf("can't close destination IDs: %x", err)
    }

    return nil
}

// /proc/self/ns/net, /sys/fs/bpf/xxx_dispatcher/
func openNetNS(nsPath, bpfFsPath string) (ns.NetNS, string, error) {
    var fs unix.Statfs_t
    err := unix.Statfs(bpfFsPath, &fs)
    if err != nil || fs.Type != unix.BPF_FS_MAGIC {
        return nil, "", fmt.Errorf("invalid BPF filesystem path: %s", bpfFsPath)
    }

    netNs, err := ns.GetNS(nsPath)
    if err != nil {
        return nil, "", err
    }

    var stat unix.Stat_t
    if err := unix.Fstat(int(netNs.Fd()), &stat); err != nil {
        return nil, "", fmt.Errorf("stat netns: %s", err)
    }

    dir := fmt.Sprintf("%d_dispatcher", stat.Ino)
    return netNs, filepath.Join(bpfFsPath, dir), nil
}
