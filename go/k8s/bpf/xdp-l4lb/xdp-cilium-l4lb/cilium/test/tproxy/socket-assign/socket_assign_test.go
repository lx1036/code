package main

import (
    "github.com/cilium/ebpf"
    "github.com/containernetworking/plugins/pkg/ns"
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "golang.org/x/sys/unix"
    tc_program_attach "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/test/tc-program-attach"
    "net"
    "os"
    "os/signal"
    "syscall"
    "testing"
    "unsafe"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go bpf test_sk_assign.c -- -I.

// go generate .

const (
    /* External (address, port) pairs the client sends packets to. */
    EXT_IP4  = "127.0.0.1"
    EXT_PORT = 4321

    /* Internal (address, port) pairs the server listens/receives at. */
    INT_IP4  = "127.0.0.2"
    INT_PORT = 1234
)

func init() {
    logrus.SetReportCaller(true)
}

type SocketAssignSuite struct {
    suite.Suite

    netns   ns.NetNS
    pinPath string
    objs    bpfObjects

    program *tc_program_attach.TcProgram
}

func (s *SocketAssignSuite) SetupSuite() {
    // 1.Load pre-compiled programs and maps into the kernel.
    objs := bpfObjects{}
    opts := &ebpf.CollectionOptions{
        Programs: ebpf.ProgramOptions{
            LogLevel: ebpf.LogLevelInstruction,
            LogSize:  64 * 1024 * 1024, // 64M
        },
        Maps: ebpf.MapOptions{
            //PinPath: pinPath, // pin 下 map, 不过 bpf 里没有定义，这里不会起作用
        },
    }
    err := loadBpfObjects(&objs, opts)
    assert.NoError(s.T(), err)
    s.objs = objs

    // attach tc ingress to lo
    info, err := s.objs.bpfPrograms.BpfSkAssignTest.Info()
    if err != nil {
        logrus.Errorf("program Entry Info err: %v", err)
        return
    }
    attachParams := &tc_program_attach.TcAttachParams{
        Interface:      "lo",
        ProgramName:    info.Name, // sk_assign_test
        ProgramFd:      s.objs.bpfPrograms.BpfSkAssignTest.FD(),
        Direction:      tc_program_attach.TcDirectionIngress,
        DirectAction:   true,
        ClobberIngress: true,
    }
    program := tc_program_attach.NewTcSchedClsProgram()
    err = program.Attach(attachParams)
    assert.NoError(s.T(), err)

    s.program = program
}

func (s *SocketAssignSuite) TearDownSuite() {
    // debug
    stopCh := make(chan os.Signal, 1)
    signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)
    <-stopCh

    s.program.Detach()

    s.objs.Close()
}

func TestSocketAssignSuite(t *testing.T) {
    suite.Run(t, new(SocketAssignSuite))
}

type ListenAt struct {
    ip   string
    port int
}

type ConnectTo struct {
    ip   string
    port int
}

// CGO_ENABLED=0 go test -v -testify.m ^TestSocketAssign$ .
// `netstat -tulpn`
func (s *SocketAssignSuite) TestSocketAssign() {
    fixtures := []struct {
        name        string
        description string
        addr        string

        progName          string
        reuseportProgName string
        socketType        int
        acceptOn          int
        reuseportHasConns bool
        listenAt          ListenAt
        connectTo         ConnectTo
    }{
        {
            name:        "ipv4 tcp port redir",
            description: "127.0.0.1:4321 > 127.0.0.1:1234",
            socketType:  unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   EXT_IP4,
                port: INT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
        },
        //{
        //    name: "ipv4 tcp addr redir",
        //},
        //{
        //    name: "ipv4 udp port redir",
        //},
        //{
        //    name: "ipv4 udp addr redir",
        //},
    }

    for _, fixture := range fixtures {
        s.T().Run(fixture.name, func(t *testing.T) {
            serverFd, err := makeServer(fixture.socketType, fixture.listenAt.ip, fixture.listenAt.port)
            if err != nil {
                logrus.Errorf("makeServer err: %v", err)
                return
            }
            defer unix.Close(serverFd)

            key := 0
            value := uint64(serverFd)
            err = s.objs.bpfMaps.ServerMap.Put(unsafe.Pointer(&key), value) // 0: server_fd
            if err != nil {
                logrus.Errorf("ServerMap.Put err: %v", err)
                return
            }

            clientFd := makeClient(fixture.socketType, fixture.connectTo.ip, fixture.connectTo.port)
            defer unix.Close(clientFd)

            var clientServerFd int
            if fixture.socketType == unix.SOCK_STREAM {
                clientServerFd, _, err = unix.Accept(serverFd)
                if err != nil {
                    logrus.Errorf("unix.Accept err: %v", err)
                    return
                }
            } else {
                clientServerFd = serverFd
            }

            buf := []byte("testing")
            _, err = unix.Write(clientFd, buf)
            if err != nil {
                logrus.Errorf("unix.Write err: %v", err)
                return
            }

            buf2 := make([]byte, 1024)
            if fixture.socketType == unix.SOCK_STREAM {
                n2, err := unix.Read(clientServerFd, buf2)
                if err != nil {
                    logrus.Errorf("unix.Read err: %v", err)
                    return
                }
                buf2 = buf2[:n2]
            } else {
                n2, _, err := unix.Recvfrom(clientServerFd, buf2, 0)
                if err != nil {
                    logrus.Errorf("unix.Recvfrom err: %v", err)
                    return
                }
                buf2 = buf2[:n2]
            }
            logrus.Infof("server unix.Recvmsg from client: %s", string(buf2))

            sa, err := unix.Getsockname(clientServerFd)
            if err != nil {
                logrus.Errorf("unix.Getsockname err: %v", err)
                return
            }

            port := sa.(*unix.SockaddrInet4).Port

            /* SOCK_STREAM is connected via accept(), so the server's local address
             * will be the CONNECT_PORT rather than the BIND port that corresponds
             * to the listen socket.
             * SOCK_DGRAM on the other hand is connectionless
             * so we can't really do the same check there; the server doesn't ever
             * create a socket with CONNECT_PORT.
             */
            if fixture.socketType == unix.SOCK_STREAM {
                if port != EXT_PORT {

                }
            } else if fixture.socketType == unix.SOCK_DGRAM {
                if port != INT_PORT {

                }
            }
        })
    }
}

func makeClient(socketType int, ip string, port int) int {
    var err error
    var sockfd int
    defer func() {
        if err != nil && sockfd > 0 {
            unix.Close(sockfd)
        }
    }()

    sockfd, err = unix.Socket(unix.AF_INET, socketType, 0)
    if err != nil {
        logrus.Fatal(err)
    }
    err = unix.SetsockoptInt(sockfd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
    if err != nil {
        logrus.Errorf("unix.SO_REUSEADDR error: %v", err)
        return 0
    }

    // client bind ip:port
    // bind ip 单元测试时会报错 "address already in use"
    ip1 := net.ParseIP(EXT_IP4)
    sa1 := &unix.SockaddrInet4{
        Port: 5432, // client 源端口
        Addr: [4]byte{},
    }
    copy(sa1.Addr[:], ip1)
    err = unix.Bind(sockfd, sa1)
    if err != nil {
        logrus.Fatal(err)
    }

    setSocketTimeout(sockfd, 1000)
    ipAddr := net.ParseIP(ip)
    sa := &unix.SockaddrInet4{
        Port: port,
        Addr: [4]byte{},
    }
    copy(sa.Addr[:], ipAddr)
    // 非阻塞的, tcp/udp 都 connect()
    err = unix.Connect(sockfd, sa)
    if err != nil {
        logrus.Errorf("unix.Connect error: %v", err)
        return 0
    }

    return sockfd
}
