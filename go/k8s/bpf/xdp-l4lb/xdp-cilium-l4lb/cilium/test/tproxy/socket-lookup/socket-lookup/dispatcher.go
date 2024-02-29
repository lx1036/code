package main

import (
    "encoding/binary"
    "fmt"
    "github.com/cilium/ebpf/link"
    "github.com/containernetworking/plugins/pkg/ns"
    "golang.org/x/sys/unix"
    "path/filepath"
    "syscall"

    "github.com/cilium/ebpf"
)

type Dispatcher struct {
    bindings     *ebpf.Map
    destinations *Destinations
}

func CreateDispatcher() (*Dispatcher, error) {
    netnsPath := "/proc/self/ns/net"
    bpfFsPath := "/sys/fs/bpf"
    netns, pinPath, err := openNetNS(netnsPath, bpfFsPath)

    objs := dispatcherObjects{}
    opts := &ebpf.CollectionOptions{
        Programs: ebpf.ProgramOptions{
            LogLevel: ebpf.LogLevelInstruction,
            LogSize:  64 * 1024 * 1024, // 64M
        },
        Maps: ebpf.MapOptions{
            PinPath: pinPath, // pin 下 map
        },
    }
    spec, err := loadDispatcher()

    var specs dispatcherSpecs
    err = spec.Assign(&specs)

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
    specs.Destinations.ValueSize = uint32(binary.Size(destinationAlloc{}))
    err = spec.LoadAndAssign(objs, opts)
    // pin program
    if err = objs.dispatcherPrograms.Dispatcher.Pin(fmt.Sprintf("%s/dispatcher_program", pinPath)); err != nil {

    }

    // Attach
    l, err := link.AttachNetNs(int(netns.Fd()), objs.dispatcherPrograms.Dispatcher)
    if err != nil {
        return nil, fmt.Errorf("attach program to netns %s: %s", netns.Path(), err)
    }
    defer l.Close()
    // pin link
    if err = l.Pin(fmt.Sprintf("%s/dispatcher_link", pinPath)); err != nil {

    }

    return &Dispatcher{
        bindings:     objs.dispatcherMaps.Bindings,
        destinations: NewDestinations(objs.dispatcherMaps),
    }, nil
}

func (dispatcher *Dispatcher) RegisterSocket(label string, conn syscall.Conn) (dest *Destination, created bool, err error) {
    dest, err := newDestinationFromConn(label, conn)
    if err != nil {
        return nil, false, err
    }

    created, err = dispatcher.destinations.AddSocket(dest, conn)
    if err != nil {
        return nil, false, fmt.Errorf("add socket: %s", err)
    }

    return
}

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

func OpenDispatcher() (*Dispatcher, error) {

}
