package main

import (
    "fmt"
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
    "os"
    "os/signal"
    "syscall"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go bpf test_udp_sk_limit.c -- -I.

const (
    CgroupPath = "/sys/fs/cgroup/udp_limit"

    PinPath = "/sys/fs/bpf/socket_service/udp_limit"
)

// go generate .
// CGO_ENABLED=0 go run .
func main() {
    logrus.SetReportCaller(true)

    stopCh := make(chan os.Signal, 1)
    signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

    joinCgroup()

    //1.Load pre-compiled programs and maps into the kernel.
    objs := bpfObjects{}
    opts := &ebpf.CollectionOptions{
        Programs: ebpf.ProgramOptions{
            LogLevel: ebpf.LogLevelInstruction,
            LogSize:  64 * 1024 * 1024, // 64M
        },
        Maps: ebpf.MapOptions{
            PinPath: PinPath, // pin 下 map
        },
    }
    err := loadBpfObjects(&objs, opts)
    if err != nil {
        logrus.Errorf("loadBpfObjects err: %v", err)
        return
    }
    defer objs.Close()

    //2.Attach programs to cgroup
    l1, err := link.AttachCgroup(link.CgroupOptions{
        Path:    CgroupPath,
        Program: objs.bpfPrograms.Sock,
        Attach:  ebpf.AttachCGroupInetSockCreate,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l1.Close()

    l2, err := link.AttachCgroup(link.CgroupOptions{
        Path:    CgroupPath,
        Program: objs.bpfPrograms.SockRelease,
        Attach:  ebpf.AttachCgroupInetSockRelease,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l2.Close()

    sk1, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
    if err != nil {
        logrus.Fatal(err)
    }

    sk2, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
    if err != nil {
        logrus.Infof("expected err: %v for sk: %d", err, sk2)
    } else {
        logrus.Fatalf("expect an error, but no error for sk: %d", sk2)
    }

    err = unix.Close(sk1)
    if err != nil {
        logrus.Fatal(err)
    }

    sk1, err = unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
    if err != nil {
        logrus.Fatal(err)
    }
    if sk1 < 0 {
        logrus.Fatalf("fail to create socket for sk1:%d", sk1)
    }

    // TODO: userspace 里获取 bpf 里的 static volatile 变量，不过还没有成功
    //  bpf_trace_printk: create socket in_use: 1, invocations: 4
    /*if skel.bss.invocations != 4 || skel.bss.in_use != 1 {
        logrus.Fatalf("error")
    }*/

    // Wait
    <-stopCh
}

// 把当前进程 pid 写到新建的 connect_force_port cgroup
func joinCgroup() {
    if err := os.MkdirAll(CgroupPath, 0777); err != nil {
        logrus.Fatalf("os.Mkdir err: %v", err)
        return
    }
    pid := os.Getpid()
    file := fmt.Sprintf("%s/cgroup.procs", CgroupPath)
    if err := os.WriteFile(file, []byte(fmt.Sprintf("%d\n", pid)), 0644); err != nil {
        logrus.Fatalf("os.WriteFile err: %v", err)
        return
    }
}

func cleanupCgroup() {
    os.RemoveAll(CgroupPath)
}
