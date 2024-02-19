package main

import (
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
    "net"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go bpf test_sk_lookup.c -- -I.

const (
    SOMAXCONN = 4096

    /*MAX_SERVERS enum {
        SERVER_A = 0,
        SERVER_B,
    };*/
    MAX_SERVERS = 2

    INT_IP4  = "127.0.0.2"
    INT_PORT = 8008
    EXT_IP4  = "127.0.0.1"
    EXT_PORT = 7007
)

const (
    ServerA = iota
    ServerB
)

// go generate .
// CGO_ENABLED=0 go run .
func main() {
    logrus.SetReportCaller(true)

    netnsPath := "/proc/self/ns/net"
    bpfFsPath := "/sys/fs/bpf"
    netns, pinPath, err := openNetNS(netnsPath, bpfFsPath)
    if err != nil {
        logrus.Errorf("openNetNS err: %v", err)
        return
    }
    defer netns.Close()

    // 1.Load pre-compiled programs and maps into the kernel.
    objs := bpfObjects{}
    opts := &ebpf.CollectionOptions{
        Programs: ebpf.ProgramOptions{
            LogLevel: ebpf.LogLevelInstruction,
            LogSize:  64 * 1024 * 1024, // 64M
        },
        Maps: ebpf.MapOptions{
            PinPath: pinPath, // pin 下 map
        },
    }
    err = loadBpfObjects(&objs, opts)
    if err != nil {
        logrus.Errorf("loadBpfObjects err: %v", err)
        return
    }
    defer objs.Close()

    // 2.attach sk_lookup/redir_port into netns
    l, err := link.AttachNetNs(int(netns.Fd()), objs.bpfPrograms.RedirPort)
    if err != nil {
        logrus.Errorf("attach program to netns %s: %s", netns.Path(), err)
        return
    }
    defer l.Close()

    // 3.update map
    serverFds := makeServer(unix.SOCK_DGRAM, nil)
    //serverFds := makeServer(unix.SOCK_STREAM, nil)
    defer unix.Close(serverFds[0])
    key := uint32(0)
    value := uint64(serverFds[0])
    err = objs.bpfMaps.RedirMap.Put(key, value)

    //reuseportHasConns := false
    //if reuseportHasConns {
    //
    //}

    // 4.echo server test
    clientFd := makeClient(unix.SOCK_DGRAM)
    //clientFd := makeClient(unix.SOCK_STREAM)
    defer unix.Close(clientFd)
    //tcpEcho(clientFd, serverFd)
    udpEcho(clientFd, serverFds[0])
}

// UDP 一直没调通, 貌似 bpf 程序没起作用!!!
// INFO: 有 bpf 程序，makeClient() 里不能 connect() 表示已经 established socket
func udpEcho(clientFd, serverFd int) {
    echoData := []byte("a")

    //clientSockAddr, err := unix.Getsockname(clientFd)
    //logrus.Infof("client port %d", clientSockAddr.(*unix.SockaddrInet4).Port) // 58442
    //
    //// 127.0.0.1.5432 > 127.0.0.1.7007
    //serverSockAddr, err := unix.Getsockname(serverFd)
    //if err != nil {
    //    logrus.Errorf("unix.Getsockname err: %v", err)
    //    return
    //}
    //serverPort := serverSockAddr.(*unix.SockaddrInet4).Port
    //logrus.Infof("server port %d", serverPort) // 8008

    ip := net.ParseIP(EXT_IP4)
    sa := &unix.SockaddrInet4{
        //Port: INT_PORT,
        Port: EXT_PORT,
        Addr: [4]byte{},
    }
    copy(sa.Addr[:], ip)
    err := unix.Sendto(clientFd, echoData, 0, sa) // 一旦 udp 7007
    if err != nil {
        logrus.Errorf("unix.Send err: %v", err)
        return
    }

    cbuf := make([]byte, unix.CmsgSpace(4))
    n, _, _, from, err := unix.Recvmsg(serverFd, cbuf, nil, 0)
    cbuf = cbuf[:n]
    logrus.Infof("server unix.Recvmsg from client: %s", string(cbuf))
    // 127.0.0.1.7007 > 127.0.0.1.5432
    // server->client 回 echo 包
    //newSockfd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
    ////serverSockAddr, err := unix.Getsockname(serverFd)
    //err = unix.Bind(newSockfd, from) // 这里报错已经监听该ip:port
    //if err != nil {
    //    logrus.Fatal(err)
    //}

    // 127.0.0.1.7007 > 127.0.0.1.5432
    clientPort := from.(*unix.SockaddrInet4).Port
    logrus.Infof("client port %d", clientPort) // 58442
    err = unix.Sendto(serverFd, cbuf, 0, from)
    if err != nil {
        logrus.Fatal(err)
    }

    //p1 := make([]byte, 1)
    cbuf2 := make([]byte, unix.CmsgSpace(4))
    n2, _, _, _, err := unix.Recvmsg(clientFd, cbuf2, nil, 0)
    if err != nil {
        logrus.Errorf("unix.Recvfrom clientFd err: %v", err)
        return
    }
    cbuf = cbuf[:n2]
    //_, _, err = unix.Recvfrom(clientFd, p1, 0)
    //if err != nil {
    //    logrus.Errorf("unix.Recvfrom clientFd err: %v", err)
    //    return
    //}
    logrus.Infof("client unix.Recvfrom from server: %s", string(cbuf))

    //cbuf2 := make([]byte, unix.CmsgSpace(4))
    //n2, _, _, _, err := unix.Recvmsg(clientFd, cbuf2, nil, 0)
    //if err != nil {
    //    logrus.Errorf("unix.Recvmsg clientFd err: %v", err)
    //    return
    //}
    //cbuf2 = cbuf2[:n2]
    //logrus.Infof("client recvmsg from server: %s", string(cbuf2))
    //
    //if cbuf2[0] == echoData[0] {
    //    logrus.Infof("udp echo successfully")
    //} else {
    //    logrus.Errorf("fail to udp echo")
    //}

}

// INFO: 目前验证，TCP 保持源端口，无需改造 server tproxy+IP_TRANSPARENT(https://powerdns.org/tproxydoc/tproxy.md.html) 就可以获得源端口
func tcpEcho(clientFd, serverFd int) {
    echoData := []byte("a")

    // 这里不使用 Sendto(fd int, p []byte, flags int, to Sockaddr) 原因是已经调用 Connect(server)
    err := unix.Send(clientFd, echoData, 0)
    if err != nil {
        logrus.Errorf("unix.Send err: %v", err)
        return
    }

    clientServerFd, clientSockAddr, err := unix.Accept(serverFd)
    if err != nil {
        logrus.Errorf("unix.Accept err: %v", err)
        return
    }
    // 保持源端口，无需改造 server IP_TRANSPARENT(https://powerdns.org/tproxydoc/tproxy.md.html) 就可以获得源端口
    clientPort := clientSockAddr.(*unix.SockaddrInet4).Port
    logrus.Infof("client port %d", clientPort) // 5432

    cbuf := make([]byte, unix.CmsgSpace(4))
    n, _, _, _, err := unix.Recvmsg(clientServerFd, cbuf, nil, 0)
    if err != nil {
        logrus.Errorf("unix.Recvmsg clientServerFd err: %v", err)
        return
    }
    cbuf = cbuf[:n]
    logrus.Infof("server recvmsg from client: %s", string(cbuf))
    err = unix.Send(clientServerFd, cbuf, 0)
    if err != nil {
        logrus.Errorf("unix.Send err: %v", err)
        return
    }

    cbuf2 := make([]byte, unix.CmsgSpace(4))
    n2, _, _, _, err := unix.Recvmsg(clientFd, cbuf2, nil, 0)
    if err != nil {
        logrus.Errorf("unix.Recvmsg clientFd err: %v", err)
        return
    }
    cbuf2 = cbuf2[:n2]
    logrus.Infof("client recvmsg from server: %s", string(cbuf2))

    if cbuf2[0] == echoData[0] {
        logrus.Infof("tcp echo successfully")
    } else {
        logrus.Errorf("fail to tcp echo")
    }
}

// 127.0.0.1:5432 connect 127.0.0.1:7007
func makeClient(socketType int) int {
    var err error
    var sockfd int
    defer func() {
        if err != nil && sockfd > 0 {
            unix.Close(sockfd)
        }
    }()

    sockfd, err = unix.Socket(unix.AF_INET, socketType, 0)
    // client bind ip:port
    // bind ip 单元测试时会报错 "address already in use"
    //ip1 := net.ParseIP(EXT_IP4)
    //sa1 := &unix.SockaddrInet4{
    //    Port: 5432, // client 源端口
    //    Addr: [4]byte{},
    //}
    //copy(sa1.Addr[:], ip1)
    //err = unix.Bind(sockfd, sa1)
    //if err != nil {
    //    logrus.Fatal(err)
    //}

    /*serverSockAddr, err := unix.Getsockname(serverFd)
      if err != nil {
          logrus.Fatal(err)
      }*/

    ip := net.ParseIP(EXT_IP4)
    sa := &unix.SockaddrInet4{
        //Port: INT_PORT,
        Port: EXT_PORT,
        Addr: [4]byte{},
    }
    copy(sa.Addr[:], ip)
    // 非阻塞的
    err = unix.Connect(sockfd, sa)
    if err != nil {
        logrus.Fatal(err)
    }

    return sockfd
}

// listen at 127.0.0.1:8008
func makeServer(socketType int, reuseportProg *ebpf.Program) [MAX_SERVERS]int {
    var sockfds [MAX_SERVERS]int
    for i := 0; i < MAX_SERVERS; i++ {
        var err error
        var sockfd int

        sockfd, err = unix.Socket(unix.AF_INET, socketType, 0)
        if err != nil {
            logrus.Errorf("unix.Socket error: %v", err)
            continue
        }
        if socketType == unix.SOCK_DGRAM {
            err = unix.SetsockoptInt(sockfd, unix.SOL_IP, unix.IP_RECVORIGDSTADDR, 1)
            if err != nil {
                logrus.Errorf("unix.IP_RECVORIGDSTADDR error: %v", err)
                continue
            }
        }
        if sockfd == unix.SOCK_STREAM {
            err = unix.SetsockoptInt(sockfd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
            if err != nil {
                logrus.Errorf("unix.SO_REUSEADDR error: %v", err)
                continue
            }
        }
        if reuseportProg != nil {
            err = unix.SetsockoptInt(sockfd, unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
            if err != nil {
                logrus.Errorf("unix.SO_REUSEPORT error: %v", err)
                continue
            }
        }

        ip := net.ParseIP(EXT_IP4)
        sa := &unix.SockaddrInet4{
            Port: INT_PORT,
            Addr: [4]byte{},
        }
        copy(sa.Addr[:], ip)
        err = unix.Bind(sockfd, sa)
        if err != nil {
            logrus.Errorf("unix.Bind error: %v", err)
            continue
        }
        if socketType == unix.SOCK_STREAM {
            err = unix.Listen(sockfd, SOMAXCONN)
            if err != nil {
                logrus.Errorf("unix.Listen error: %v", err)
                continue
            }
        }

        // attach reuseport program
        if reuseportProg != nil {
            err = unix.SetsockoptInt(sockfd, unix.SOL_SOCKET, unix.SO_ATTACH_REUSEPORT_EBPF, reuseportProg.FD())
            if err != nil {
                logrus.Errorf("unix.SO_ATTACH_REUSEPORT_EBPF error: %v", err)
                continue
            }
        }

        sockfds[i] = sockfd
        if reuseportProg == nil {
            break
        }
    }

    return sockfds
}
