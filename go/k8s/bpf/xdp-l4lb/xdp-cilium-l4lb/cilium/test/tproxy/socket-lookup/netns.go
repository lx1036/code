package main

import (
    "fmt"
    "github.com/containernetworking/plugins/pkg/ns"
    "golang.org/x/sys/unix"
    "path/filepath"
)

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
