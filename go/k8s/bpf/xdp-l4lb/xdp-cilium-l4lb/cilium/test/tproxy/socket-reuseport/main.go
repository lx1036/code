package main

import (
    "fmt"
    "github.com/cilium/ebpf"
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
    "net"
    "unsafe"
)

/**
测试 map type:
BPF_MAP_TYPE_REUSEPORT_SOCKARRAY
BPF_MAP_TYPE_SOCKMAP
BPF_MAP_TYPE_SOCKHASH
*/

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go bpf test_sk_reuseport.c -- -I.

const (
    REUSEPORT_ARRAY_SIZE = 32

    LOOPBACK    = "127.0.0.1"
    SERVER_PORT = 8008
    SOMAXCONN   = 4096

    PinPath = "/sys/fs/bpf/socket_reuseport"
)

var (
    sk_fds [REUSEPORT_ARRAY_SIZE]int

    epfd int

    serverSockAddr unix.Sockaddr

    reuseport_array_map *ebpf.Map
    outer_map           *ebpf.Map
)

type Cmd struct {
    reuseport_index uint32
    pass_on_failure uint32
}

// go generate .
// CGO_ENABLED=0 go run .
func main() {
    logrus.SetReportCaller(true)

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
    reuseport_array_spec := create_maps()
    spec, err := loadBpf()
    if err != nil {
        logrus.Errorf("loadBpf err: %v", err)
        return
    }
    spec.Maps["outer_map"].InnerMap = &reuseport_array_spec
    err = spec.LoadAndAssign(&objs, opts)
    if err != nil {
        logrus.Errorf("LoadAndAssign err: %v", err)
        return
    }
    defer objs.Close()

    //2.
    err = objs.bpfMaps.OuterMap.Put(uint32(0), uint32(reuseport_array_map.FD()))
    prepare_sk_fds(unix.SOCK_STREAM, objs.bpfPrograms.SelectBySkbData.FD())
    ovr := int(-1)
    err = objs.bpfMaps.IndexMap.Put(uint32(0), ovr)
    if err != nil {
        logrus.Errorf("indexMap.Put err: %v", err)
        return
    }

    //3. 验证测试
    test_pass()

}

func setup_per_test(indexMap *ebpf.Map) {
    //prepare_sk_fds(unix.SOCK_STREAM)

    ovr := int(-1)
    err := indexMap.Put(uint32(0), ovr)
    if err != nil {
        logrus.Errorf("indexMap.Put err: %v", err)
        return
    }

    //err := outer_map.Put(uint32(0), uint32(reuseport_array_map.FD()))
}

func create_maps() ebpf.MapSpec {
    var err error
    reuseport_array_spec := ebpf.MapSpec{
        Name:       "reuseport_array",
        Type:       ebpf.ReusePortSockArray, // TODO: 还需测试 SockHash, SockMap
        KeySize:    uint32(unsafe.Sizeof(uint32(0))),
        ValueSize:  uint32(unsafe.Sizeof(uint32(0))),
        MaxEntries: REUSEPORT_ARRAY_SIZE,
    }
    reuseport_array_map, err = ebpf.NewMapWithOptions(&reuseport_array_spec, ebpf.MapOptions{
        PinPath: fmt.Sprintf("%s/%s", PinPath, "reuseport_array"),
    })
    if err != nil {

    }

    //outer_map, err = ebpf.NewMapWithOptions(&ebpf.MapSpec{
    //    Name:       "outer_map",
    //    Type:       ebpf.ArrayOfMaps,
    //    KeySize:    uint32(unsafe.Sizeof(uint32(0))),
    //    ValueSize:  uint32(unsafe.Sizeof(uint32(0))),
    //    MaxEntries: 1,
    //    InnerMap:   &reuseport_array_spec,
    //}, ebpf.MapOptions{
    //    PinPath: fmt.Sprintf("%s/%s", PinPath, "outer_map"),
    //})
    //if err != nil {
    //
    //}

    return reuseport_array_spec
}

func prepare_sk_fds(socketType, reuseportProgFd int) {
    var err error

    /*
     * The sk_fds[] is filled from the back such that the order
     * is exactly opposite to the (struct sock_reuseport *)reuse->socks[].
     */
    for i := REUSEPORT_ARRAY_SIZE; i > 0; i-- {
        sk_fds[i], err = unix.Socket(unix.AF_INET, socketType, 0)

        err = unix.SetsockoptInt(sk_fds[i], unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)

        // INFO: 1.挂载 reuseport bpf program
        if i == REUSEPORT_ARRAY_SIZE {
            err = unix.SetsockoptInt(sk_fds[i], unix.SOL_SOCKET, unix.SO_ATTACH_REUSEPORT_EBPF, reuseportProgFd)
        }

        ip := net.ParseIP(LOOPBACK)
        sa := &unix.SockaddrInet4{
            Port: SERVER_PORT,
            Addr: [4]byte{},
        }
        copy(sa.Addr[:], ip)
        err = unix.Bind(sk_fds[i], sa)

        if socketType == unix.SOCK_STREAM {
            err = unix.Listen(sk_fds[i], SOMAXCONN)
            if err != nil {
                logrus.Errorf("unix.Listen error: %v", err)
                return
            }
        }

        // INFO: 这里更新 reuseport_map[i]=sk_fd
        err = reuseport_array_map.Put(i, sk_fds[i])

        if i == REUSEPORT_ARRAY_SIZE {
            serverSockAddr, err = unix.Getsockname(sk_fds[i])
        }

    }

    epfd, err = unix.EpollCreate(1)
    ev := unix.EpollEvent{
        Events: unix.EPOLLIN,
        //Fd:     sk_fds[0], // 注意这个参数
    }
    for i := 0; i < REUSEPORT_ARRAY_SIZE; i++ {
        ev.Pad = int32(i) // 这里是为了让 epoll_wait 处能够正确的获取哪个 server_fd 有读数据
        err = unix.EpollCtl(epfd, unix.EPOLL_CTL_ADD, sk_fds[i], &ev)
    }
}

func test_pass() {
    cmd := Cmd{
        reuseport_index: 0,
        pass_on_failure: 0,
    }

    do_test(unix.SOCK_STREAM, cmd)
}

func do_test(socketType int, cmd Cmd) {
    //data := [8]byte{}
    //copy(data[:], "00010001")
    data := []byte("00010001") // 必须是 8 字节
    send_data(socketType, data)

    events := make([]unix.EpollEvent, 1024)

    /**
      INFO: 注意这里 server 没有使用 listen, 而是使用了 epoll_wait(低 level api)，然后去 accept 客户端发来的数据。参考 socket-lookup/main.go
          这么做的原因是，构建一组 reuseport sk_fds，然后获取 sk_fds[i] 中的数据
    */

    // 参数 timeout 表示在没有检测到事件发生时最多等待的时间（单位为毫秒），如果 timeout为0，则表示 epoll_wait在 rdllist链表中为空，立刻返回，不会等待
    // INFO: epoll_wait 阻塞的
    _, err := unix.EpollWait(epfd, events, 5)
    if err != nil {
        logrus.Errorf("unix.EpollWait err: %v", err)
        return
    }

    check_results()
    check_data()

    index := 0
    for i, event := range events {
        if event.Pad != 0 {
            index = i
            break
        }
    }
    srvFd := sk_fds[index]

    if socketType == unix.SOCK_STREAM {
        clientServerFd, _, err := unix.Accept(srvFd)
        if err != nil {
            logrus.Errorf("unix.Accept err: %v", err)
            return
        }
        cbuf := make([]byte, unix.CmsgSpace(4))
        n, _, _, _, err := unix.Recvmsg(clientServerFd, cbuf, nil, unix.MSG_DONTWAIT)
        if err != nil {
            logrus.Errorf("unix.Recvmsg err: %v", err)
            return
        }
        cbuf = cbuf[:n]
        logrus.Infof("server unix.Recvmsg from client: %s", string(cbuf))
    } else {
        cbuf := make([]byte, unix.CmsgSpace(4))
        n, _, _, _, err := unix.Recvmsg(srvFd, cbuf, nil, 0)
        if err != nil {
            logrus.Errorf("unix.Recvmsg err: %v", err)
            return
        }
        cbuf = cbuf[:n]
        logrus.Infof("server unix.Recvmsg from client: %s", string(cbuf))
    }

}

func send_data(socketType int, data []byte) int {
    clientFd, err := unix.Socket(unix.AF_INET, socketType, 0)

    // INFO: 需要这个 socket option 么???
    err = unix.SetsockoptInt(clientFd, unix.SOL_TCP, unix.TCP_FASTOPEN, 256)
    if err != nil {
        logrus.Fatal(err)
    }

    // client bind ip:port
    // bind ip 单元测试时会报错 "address already in use"
    ip := net.ParseIP(LOOPBACK)
    sa := &unix.SockaddrInet4{
        Port: 5432, // client 源端口
        Addr: [4]byte{},
    }
    copy(sa.Addr[:], ip)
    err = unix.Bind(clientFd, sa)
    if err != nil {
        logrus.Fatal(err)
    }

    //var data2 []byte
    //copy(data2[:], data[:])
    err = unix.Sendto(clientFd, data, unix.MSG_FASTOPEN, serverSockAddr)
    if err != nil {
        logrus.Fatal(err)
    }

    return clientFd
}

func check_results() {

}

func check_data() {

}
