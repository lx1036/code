package main

// /root/linux-5.10.142/tools/testing/selftests/bpf/prog_tests/tcp_hdr_options.c

import (
    "fmt"
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
    "net"
    "os"
    "testing"
)

const (
    CgroupPath = "/sys/fs/cgroup/tcphdr"

    INADDR_LOOPBACK = "127.0.0.1"
    BindPort        = 5432
)

const (
    OPTION_RESEND = iota
    OPTION_MAX_DELACK_MS
    OPTION_RAND
    __NR_OPTION_FLAGS
)

const (
    OPTION_F_RESEND        = 1 << OPTION_RESEND
    OPTION_F_MAX_DELACK_MS = 1 << OPTION_MAX_DELACK_MS
    OPTION_F_RAND          = 1 << OPTION_RAND
    OPTION_MASK            = (1 << __NR_OPTION_FLAGS) - 1
)

type SkFds struct {
    srv_fd        int // serverFd
    active_fd     int // clientFd
    passive_fd    int // clientServerFd
    passive_lport int // serverPort
    active_lport  int // clientPort
}

func TestSimpleEstablish(test *testing.T) {
    skFds := SkFds{}

    cgroupPath := joinCgroup("tcphdr_opt")
    //defer cleanupCgroup()

    err := os.WriteFile("/proc/sys/net/ipv4/tcp_syncookies", []byte("1"), 0600)

    objs := bpfObjects{}
    opts := &ebpf.CollectionOptions{
        Programs: ebpf.ProgramOptions{
            LogLevel: ebpf.LogLevelInstruction,
            LogSize:  64 * 1024 * 1024, // 64M
        },
        Maps: ebpf.MapOptions{
            //PinPath: PinPath2, // pin 下 map
        },
    }
    spec, err := loadBpf()
    if err != nil {
        logrus.Fatal(err)
    }

    if err = spec.RewriteConstants(getRewriteConstForSimpleEstablish(true)); err != nil {
        logrus.Fatal(err)
    }
    if err := spec.LoadAndAssign(&objs, opts); err != nil {
        logrus.Fatalf("loading objects: %v", err)
    }
    defer objs.Close()

    //2.Attach programs to cgroup
    l1, err := link.AttachCgroup(link.CgroupOptions{
        Path:    cgroupPath,
        Program: objs.bpfPrograms.Estab,
        Attach:  ebpf.AttachCGroupSockOps,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l1.Close()

    err = skFdsConnect(&skFds, false)
    if err != nil {
        logrus.Error(err)
    }

    checkHdrAndCloseFds(&skFds)
}

func TestNoExperimentalEstablish(test *testing.T) {

}

func getRewriteConstForSimpleEstablish(exprm bool) map[string]interface{} {
    exp_passive_estab_in := bpfBpfTestOption{
        Flags:       OPTION_F_RAND | OPTION_F_MAX_DELACK_MS,
        MaxDelackMs: 11,
        Rand:        0xfa,
    }
    exp_active_estab_in := bpfBpfTestOption{
        Flags:       OPTION_F_RAND | OPTION_F_MAX_DELACK_MS,
        MaxDelackMs: 22,
        Rand:        0xce,
    }
    consts := map[string]interface{}{
        "active_syn_out":     exp_passive_estab_in,
        "passive_synack_out": exp_active_estab_in,
        //"active_fin_out":     exp_passive_fin_in,
        //"passive_fin_out":    exp_active_fin_in,
    }

    if !exprm {
        consts["test_kind"] = 0xB9
        consts["test_magic"] = 0
    }

    return consts
}

func skFdsConnect(skFds *SkFds, fastOpen bool) error {
    skFds.srv_fd = startServer2(unix.SOCK_STREAM)
    sa1, err := unix.Getsockname(skFds.srv_fd)
    if err != nil {
        logrus.Errorf("unix.Getsockname err: %v", err)
        return err
    }
    skFds.passive_lport = sa1.(*unix.SockaddrInet4).Port

    msg := []byte("FAST!!!")
    if fastOpen {
        skFds.active_fd = fastOpenConnectToFd(skFds.srv_fd, msg)
    } else {
        skFds.active_fd = connectToFd(skFds.srv_fd)
    }
    sa2, err := unix.Getsockname(skFds.active_fd)
    if err != nil {
        logrus.Errorf("unix.Getsockname err: %v", err)
        return err
    }
    skFds.active_lport = sa2.(*unix.SockaddrInet4).Port

    skFds.passive_fd, _, err = unix.Accept(skFds.srv_fd)
    if err != nil {
        logrus.Errorf("Accept err: %v", err)
        return err
    }

    if fastOpen {
        cbuf := make([]byte, unix.CmsgSpace(4))
        n, _, _, _, err := unix.Recvmsg(skFds.passive_fd, cbuf, nil, 0)
        if err != nil {
            logrus.Errorf("Accept err: %v", err)
            return err
        }
        cbuf = cbuf[:n]
        if string(cbuf) != string(msg) {
            logrus.Errorf("read fastopen syn data err: expected=%s actual=%s", msg, cbuf)
        } else {
            logrus.Infof("read fastopen syn data success: %s", cbuf)
        }
    }

    return nil
}

func checkHdrAndCloseFds(skFds *SkFds) {
    skFdsShutdown(skFds)

    // check
    // TODO: 如何在 userspace access bpf 里的 volatile 变量

    skFdsClose(skFds)
}

func skFdsClose(skFds *SkFds) {
    unix.Close(skFds.srv_fd)
    unix.Close(skFds.passive_fd)
    unix.Close(skFds.active_fd)
}

func skFdsShutdown(skFds *SkFds) {
    aByte := make([]byte, 1)
    unix.Shutdown(skFds.active_fd, unix.SHUT_WR) // clientFd
    _, err := unix.Read(skFds.passive_fd, aByte) // server read a byte
    if err != nil {
        logrus.Errorf("read-after-shutdown(passive_fd): %v", err)
        return
    }
    logrus.Infof("a byte: %s", aByte)

    unix.Shutdown(skFds.passive_fd, unix.SHUT_WR)
    _, err = unix.Read(skFds.active_fd, aByte) // client read a byte
    if err != nil {
        logrus.Errorf("read-after-shutdown(active_fd): %v", err)
        return
    }
    logrus.Infof("a byte: %s", aByte)
}

func startServer2(sockType int) int {
    var err error
    var fd int
    defer func() {
        if err != nil && fd > 0 {
            unix.Close(fd)
        }
    }()

    fd, err = unix.Socket(unix.AF_INET, sockType, 0)
    if err != nil {
        logrus.Fatal(err)
    }

    setSocketTimeout(fd, 5000)

    // INFO: 解决报错 "address already in use", 因为 client/server tcp connection 没有正常关闭，会导致 server tcp 状态机进入 TIME_WAIT
    //  状态，需要等待 2 * Maximum Segment Lifetime=4min 时间后才能释放，测试也不是4min，大概几十秒
    //  `netstat | grep 1234`
    //  tcp        0      0 localhost:postgresql    localhost:1234          TIME_WAIT
    // ignores TIME-WAIT state using SO_REUSEADDR option
    // https://serverfault.com/questions/329845/how-to-forcibly-close-a-socket-in-time-wait
    err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
    if err != nil {
        logrus.Fatalf("unix.SO_REUSEADDR error: %v", err)
    }

    ip := net.ParseIP(INADDR_LOOPBACK)
    sa := &unix.SockaddrInet4{
        Port: BindPort,
        Addr: [4]byte{},
    }
    copy(sa.Addr[:], ip)
    err = unix.Bind(fd, sa)
    if err != nil {
        logrus.Fatal(err)
    }

    if sockType == unix.SOCK_STREAM {
        err = unix.Listen(fd, 1)
        if err != nil {
            logrus.Fatal(err)
        }
    }

    return fd
}

func connectToFd(serverFd int) int {
    socketType, err := unix.GetsockoptInt(serverFd, unix.SOL_SOCKET, unix.SO_TYPE)
    if err != nil {
        logrus.Fatal(err)
    }

    serverSockAddr, err := unix.Getsockname(serverFd)
    if err != nil {
        logrus.Fatal(err)
    }

    clientFd, err := unix.Socket(unix.AF_INET, socketType, 0)
    if err != nil {
        logrus.Fatal(err)
    }
    setSocketTimeout(clientFd, 5000)

    // 非阻塞的
    err = unix.Connect(clientFd, serverSockAddr)
    if err != nil {
        logrus.Fatal(err)
    }

    return clientFd
}

func fastOpenConnectToFd(serverFd int, msg []byte) int {
    clientFd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
    if err != nil {
        logrus.Fatal(err)
    }
    setSocketTimeout(clientFd, 5000)

    // INFO: 需要这个 socket option 么???
    err = unix.SetsockoptInt(clientFd, unix.SOL_TCP, unix.TCP_FASTOPEN, 256)
    if err != nil {
        logrus.Fatal(err)
    }

    serverSockAddr, err := unix.Getsockname(serverFd)
    if err != nil {
        logrus.Fatal(err)
    }
    err = unix.Sendto(clientFd, msg, unix.MSG_FASTOPEN, serverSockAddr)
    if err != nil {
        logrus.Fatal(err)
    }

    return clientFd
}

func setSocketTimeout(fd, timeoutMs int) {
    var timeVal *unix.Timeval
    if timeoutMs > 0 {
        timeVal = &unix.Timeval{
            Sec:  int64(timeoutMs / 1000),
            Usec: int64(timeoutMs % 1000 * 1000),
        }
    } else {
        timeVal = &unix.Timeval{
            Sec: 3,
        }
    }

    err := unix.SetsockoptTimeval(fd, unix.SOL_SOCKET, unix.SO_RCVTIMEO, timeVal)
    if err != nil {
        logrus.Fatal(err)
    }

    err = unix.SetsockoptTimeval(fd, unix.SOL_SOCKET, unix.SO_SNDTIMEO, timeVal)
    if err != nil {
        logrus.Fatal(err)
    }
}

// 把当前进程 pid 写到新建的 connect_force_port cgroup
func joinCgroup(path string) string {
    if len(path) == 0 {
        logrus.Fatalf("path is empty")
    }

    cgroupPath := fmt.Sprintf("%s/%s", CgroupPath, path)
    if err := os.MkdirAll(cgroupPath, 0777); err != nil {
        logrus.Fatalf("os.Mkdir err: %v", err)
    }
    pid := os.Getpid()
    file := fmt.Sprintf("%s/cgroup.procs", cgroupPath)
    if err := os.WriteFile(file, []byte(fmt.Sprintf("%d\n", pid)), 0644); err != nil {
        logrus.Fatalf("os.WriteFile err: %v", err)
    }

    return cgroupPath
}

func cleanupCgroup() {
    os.RemoveAll(CgroupPath)
}
