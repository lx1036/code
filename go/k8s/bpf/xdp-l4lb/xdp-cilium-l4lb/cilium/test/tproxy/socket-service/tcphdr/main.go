package main

import (
    "errors"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "syscall"

    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go bpf test_tcp_hdr_options.c -- -I.

// go generate .
// CGO_ENABLED=0 go run .
func main() {

}

type TcpHdr struct {
    stateDir *File
    Path     string
    // bindings     *ebpf.Map
    // destinations *destinations
}

// "/proc/self/ns/net": 当前进程的网络命名空间
// "/sys/fs/bpf"
func CreateTcpHdr(netnsPath, bpfFsPath string) (_ *TcpHdr, err error) {
    closeOnError := func(c io.Closer) {
        if err != nil {
            c.Close()
        }
    }

    // netns(/proc/self/ns/net), /sys/fs/bpf/{inode}_tcp_hdr
    netns, pinPath, err := openNetNS(netnsPath, bpfFsPath)
    if err != nil {
        return nil, err
    }
    defer netns.Close()

    tempDir, err := os.MkdirTemp(filepath.Dir(pinPath), "tcp-hdr-*")
    if err != nil {
        return nil, fmt.Errorf("can't create temp directory: %s", err)
    }
    defer os.RemoveAll(tempDir)

    stateDir, err := OpenLockedExclusive(tempDir)
    if err != nil {
        return nil, err
    }
    defer closeOnError(stateDir)

    var objs tcpHeaderObjects
    _, err = loadPatchedTcpHeader(&objs, &ebpf.CollectionOptions{
        Maps: ebpf.MapOptions{PinPath: tempDir},
    })
    if err != nil {
        return nil, fmt.Errorf("load BPF: %s", err)
    }
    defer objs.tcpHeaderPrograms.Close()
    defer closeOnError(&objs.tcpHeaderMaps)

    // pin /sys/fs/bpf/tcp-hdr-*/program
    if err := objs.tcpHeaderPrograms.Estab.Pin(programPath(tempDir)); err != nil {
        return nil, fmt.Errorf("pin program: %s", err)
    }

    // attach "sockops/estab" prog to a network ns
    l, err := link.AttachNetNs(int(netns.Fd()), objs.tcpHeaderPrograms.Estab)
    if err != nil {
        return nil, fmt.Errorf("attach program to netns %s: %s", netns.Path(), err)
    }
    defer l.Close()
    if err := l.Pin(linkPath(tempDir)); err != nil { // 这里的 link 去 pin，什么意思???
        return nil, fmt.Errorf("can't pin link: %s", err)
    }

    if err := adjustPermissions(tempDir); err != nil {
        return nil, fmt.Errorf("adjust permissions: %s", err)
    }

    // Rename will succeed if pinPath doesn't exist or is an empty directory,
    // otherwise it will return an error. In that case tempDir is removed,
    // and the pinned link + program are closed, undoing any changes.
    if err := os.Rename(tempDir, pinPath); os.IsExist(err) || errors.Is(err, syscall.ENOTEMPTY) {
        return nil, fmt.Errorf("can't create dispatcher: %v", err)
    } else if err != nil {
        return nil, fmt.Errorf("can't create dispatcher: %s", err)
    }

    return &TcpHdr{
        stateDir: stateDir,
        Path:     pinPath,
    }, nil
}

func adjustPermissions(path string) error {
    const (
        // Only let group list and open the directory. This is important since
        // being able to open a directory implies being able to flock it.
        dirMode os.FileMode = 0750
        // Allow group read-only access to state.
        objMode os.FileMode = 0640
    )

    if err := os.Chmod(path, dirMode); err != nil {
        return err
    }

    entries, err := os.ReadDir(path)
    if err != nil {
        return fmt.Errorf("read state entries: %s", err)
    }

    for _, entry := range entries {
        if entry.IsDir() {
            return fmt.Errorf("change access mode: %q is a directory", entry.Name())
        }

        path := filepath.Join(path, entry.Name())
        if err := os.Chmod(path, objMode); err != nil {
            return err
        }
    }

    return nil
}

func loadPatchedTcpHeader(to interface{}, opts *ebpf.CollectionOptions) (*ebpf.CollectionSpec, error) {
    spec, err := loadTcpHeader()
    if err != nil {
        return nil, err
    }

    // before loaded into kernel
    var specs tcpHeaderSpecs
    if err = spec.Assign(&specs); err != nil {
        return nil, err
    }

    maxLinum := specs.LportLinumMap.MaxEntries
    for _, m := range []*ebpf.MapSpec{
        specs.LportLinumMap,
    } {
        if m.MaxEntries != maxLinum {
            return nil, fmt.Errorf("map %q has %d max entries instead of %d", m.Name, m.MaxEntries, maxLinum)
        }
    }
    // specs.LportLinumMap.KeySize = uint32(binary.Size(int))

    if to != nil {
        return spec, spec.LoadAndAssign(to, opts)
    }

    return spec, nil
}
