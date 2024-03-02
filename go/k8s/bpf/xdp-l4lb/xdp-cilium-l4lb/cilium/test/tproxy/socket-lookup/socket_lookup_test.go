package main

import (
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
    "github.com/containernetworking/plugins/pkg/ns"
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "golang.org/x/sys/unix"
    "os"
    "testing"
    "unsafe"
)

/**
/root/linux-5.10.142/tools/testing/selftests/bpf/prog_tests/sk_lookup.c
*/

func init() {
    logrus.SetReportCaller(true)
}

func TestSocketLookupSuite(t *testing.T) {
    suite.Run(t, new(SocketLookupSuite))
}

type SocketLookupSuite struct {
    suite.Suite

    netns   ns.NetNS
    pinPath string
    objs    bpfObjects
}

func (s *SocketLookupSuite) SetupSuite() {
    logrus.SetReportCaller(true)

    netnsPath := "/proc/self/ns/net"
    bpfFsPath := "/sys/fs/bpf"
    netns, pinPath, err := openNetNS(netnsPath, bpfFsPath)
    assert.NoError(s.T(), err)
    s.netns = netns
    s.pinPath = pinPath

    // 1.Load pre-compiled programs and maps into the kernel.
    objs := bpfObjects{}
    opts := &ebpf.CollectionOptions{
        Programs: ebpf.ProgramOptions{
            LogLevel: ebpf.LogLevelInstruction,
            LogSize:  64 * 1024 * 1024, // 64M
        },
        Maps: ebpf.MapOptions{
            PinPath: pinPath, // pin 下 map, 不过 bpf 里没有定义，这里不会起作用
        },
    }
    err = loadBpfObjects(&objs, opts)
    assert.NoError(s.T(), err)
    s.objs = objs
}

func (s *SocketLookupSuite) TearDownSuite() {
    s.netns.Close()
    s.objs.Close()

    // unload 卸载 pin 的 map/program/link
    os.RemoveAll(s.pinPath)
}

func (s *SocketLookupSuite) SetupTest() {
}

func (s *SocketLookupSuite) TearDownTest() {
}

type ListenAt struct {
    ip   string
    port int
}

type ConnectTo struct {
    ip   string
    port int
}

// CGO_ENABLED=0 go test -v -testify.m ^TestSocketLookupWithReuseport$ .
// `netstat -tulpn`
func (s *SocketLookupSuite) TestSocketLookupWithReuseport() {
    fixtures := []struct {
        name              string
        description       string
        progName          string
        reuseportProgName string
        socketType        int
        acceptOn          int
        reuseportHasConns bool
        listenAt          ListenAt
        connectTo         ConnectTo
    }{
        // test_redirect_lookup()
        {
            name:        "TCP IPv4 redir port",
            description: "127.0.0.1:7007 > 127.0.0.1:8008",
            progName:    "redir_port",
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
        {
            name:        "TCP IPv4 redir addr",
            description: "127.0.0.1:7007 > 127.0.0.2:7007",
            progName:    "redir_ip4",
            socketType:  unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   INT_IP4,
                port: EXT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
        },
        {
            // INFO: reuseport 这块还是没想明白???
            name:              "TCP IPv4 redir with reuseport",
            description:       "127.0.0.1:7007 > 127.0.0.2:8008(KEY_SERVER_A, KEY_SERVER_B)，会 lookup KEY_SERVER_B",
            progName:          "select_sock_a",
            reuseportProgName: "select_sock_b",
            socketType:        unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   INT_IP4,
                port: INT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            acceptOn: ServerB,
        },
        {
            name:              "TCP IPv4 redir skip reuseport",
            description:       "127.0.0.1:7007 > 127.0.0.2:8008(KEY_SERVER_A, KEY_SERVER_B)，会 lookup KEY_SERVER_A",
            progName:          "select_sock_a_no_reuseport",
            reuseportProgName: "select_sock_b",
            socketType:        unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   INT_IP4,
                port: INT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            acceptOn: ServerA,
        },

        {
            name:        "UDP IPv4 redir port",
            description: "127.0.0.1:7007 > 127.0.0.1:8008",
            progName:    "redir_port",
            socketType:  unix.SOCK_DGRAM,
            listenAt: ListenAt{
                ip:   EXT_IP4,
                port: INT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
        },
        {
            name:        "UDP IPv4 redir addr",
            description: "127.0.0.1:7007 > 127.0.0.2:7007",
            progName:    "redir_ip4",
            socketType:  unix.SOCK_DGRAM,
            listenAt: ListenAt{
                ip:   INT_IP4,
                port: EXT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
        },
        {
            name:              "UDP IPv4 redir with reuseport",
            description:       "127.0.0.1:7007 > 127.0.0.2:7007",
            progName:          "select_sock_a",
            reuseportProgName: "select_sock_b",
            socketType:        unix.SOCK_DGRAM,
            listenAt: ListenAt{
                ip:   INT_IP4,
                port: EXT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            acceptOn: ServerB, /* ServerA */
        },
        {
            name:              "UDP IPv4 redir and reuseport with conns",
            description:       "127.0.0.1:7007 > 127.0.0.2:8008",
            progName:          "select_sock_a",
            reuseportProgName: "select_sock_b",
            socketType:        unix.SOCK_DGRAM,
            listenAt: ListenAt{
                ip:   INT_IP4,
                port: INT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            acceptOn:          ServerB, /* ServerA */
            reuseportHasConns: true,
        },
        {
            name:              "UDP IPv4 redir skip reuseport",
            progName:          "select_sock_a_no_reuseport",
            reuseportProgName: "select_sock_b",
            socketType:        unix.SOCK_DGRAM,
            listenAt: ListenAt{
                ip:   INT_IP4,
                port: INT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            acceptOn: ServerA,
        },

        // test_drop_on_lookup()
        {
            name:       "TCP IPv4 drop on lookup",
            progName:   "lookup_drop",
            socketType: unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
        },
        {
            name:       "UDP IPv4 drop on lookup",
            progName:   "lookup_drop",
            socketType: unix.SOCK_DGRAM,
            listenAt: ListenAt{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
        },

        // test_drop_on_reuseport()
        {
            name:              "TCP IPv4 drop on reuseport",
            progName:          "select_sock_a",
            reuseportProgName: "reuseport_drop",
            socketType:        unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   INT_IP4,
                port: INT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
        },
        {
            name:              "UDP IPv4 drop on reuseport",
            progName:          "select_sock_a",
            reuseportProgName: "reuseport_drop",
            socketType:        unix.SOCK_DGRAM,
            listenAt: ListenAt{
                ip:   INT_IP4,
                port: INT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
        },

        // test_sk_assign_helper()
        {
            name:        "sk_assign returns EEXIST",
            description: "127.0.0.1:7007 > 127.0.0.2:8008",
            progName:    "sk_assign_eexist",
            socketType:  unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   INT_IP4,
                port: INT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            acceptOn: ServerB, /* ServerA */
        },
        {
            name:        "sk_assign honors F_REPLACE",
            description: "127.0.0.1:7007 > 127.0.0.2:8008",
            progName:    "sk_assign_replace_flag",
            socketType:  unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   INT_IP4,
                port: INT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            acceptOn: ServerB, /* ServerA */
        },
        {
            name:        "sk_assign accepts NULL socket",
            description: "127.0.0.1:7007 > 127.0.0.2:8008",
            progName:    "sk_assign_null",
            socketType:  unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   INT_IP4,
                port: INT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            acceptOn: ServerB, /* ServerA */
        },
        {
            name:        "access ctx->sk",
            description: "127.0.0.1:7007 > 127.0.0.2:8008",
            progName:    "access_ctx_sk",
            socketType:  unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   INT_IP4,
                port: INT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            acceptOn: ServerB, /* ServerA */
        },
        {
            name:        "narrow access to ctx v4",
            description: "127.0.0.1:7007 > 127.0.0.2:8008",
            progName:    "ctx_narrow_access",
            socketType:  unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   INT_IP4,
                port: INT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            acceptOn: ServerB, /* ServerA */
        },
    }
    for _, fixture := range fixtures {
        s.T().Run(fixture.name, func(t *testing.T) {
            // 2.attach sk_lookup/redir_port into netns, 注意 reuseport bpf 程序的 attach 是直接 attach socketFd
            l, err := link.AttachNetNs(int(s.netns.Fd()), s.getBpfProg(fixture.progName))
            assert.NoError(s.T(), err)
            defer l.Close()

            // 3.update map
            serverFds := make([]int, MAX_SERVERS)
            for i := 0; i < MAX_SERVERS; i++ {
                serverFd, err := makeServer(fixture.socketType, s.getBpfProg(fixture.reuseportProgName), fixture.listenAt.ip, fixture.listenAt.port)
                if err != nil {
                    continue
                }
                serverFds[i] = serverFd
                if len(fixture.reuseportProgName) == 0 {
                    break
                }
            }
            defer func() {
                for i := 0; i < len(serverFds); i++ {
                    unix.Close(serverFds[i])
                }
            }()

            for i := 0; i < len(serverFds); i++ {
                key := uint32(i)
                value := uint64(serverFds[i])
                err = s.objs.bpfMaps.RedirMap.Put(key, value)
            }

            /* Regular UDP socket lookup with reuseport behaves
             * differently when reuseport group contains connected
             * sockets. Check that adding a connected UDP socket to the
             * reuseport group does not affect how reuseport works with
             * BPF socket lookup.
             */
            if fixture.reuseportHasConns {
                /* Add an extra socket to reuseport group */
                reuseConnFd, err := makeServer(fixture.socketType, s.getBpfProg(fixture.reuseportProgName), fixture.listenAt.ip, fixture.listenAt.port)
                if err != nil {
                    return
                }
                defer unix.Close(reuseConnFd)

                // INFO: Connect the extra socket to itself, 注意还可以 connect socket itself
                reuseConnSockAddr, err := unix.Getsockname(reuseConnFd)
                assert.NoError(s.T(), err)
                err = unix.Connect(reuseConnFd, reuseConnSockAddr)
                assert.NoError(s.T(), err)
            }

            // 4.tcp/udp echo
            clientFd := makeClient(fixture.socketType, fixture.connectTo.ip, fixture.connectTo.port)
            defer unix.Close(clientFd)

            if fixture.socketType == unix.SOCK_STREAM {
                tcpEcho(clientFd, serverFds[fixture.acceptOn], getEchoData(fixture.acceptOn))
            } else {
                udpEcho(clientFd, serverFds[fixture.acceptOn], getEchoData(fixture.acceptOn))
            }
        })
    }

    fixtures2 := []struct {
        name              string
        description       string
        progName          string
        reuseportProgName string
        socketType        int
        acceptOn          int
        reuseportHasConns bool
        listenAt          ListenAt
        connectTo         ConnectTo
    }{
        // test_redirect_lookup(), bpf map 里 connected sk, 无法 bpf_sk_assign()
        {
            name:        "sk_assign rejects TCP established",
            description: "127.0.0.1:7007 > 127.0.0.1:8008",
            progName:    "sk_assign_estabsocknosupport",
            socketType:  unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            acceptOn: ServerA,
        },
        {
            name:        "sk_assign rejects UDP connected",
            description: "127.0.0.1:7007 > 127.0.0.1:8008",
            progName:    "sk_assign_estabsocknosupport",
            socketType:  unix.SOCK_DGRAM,
            listenAt: ListenAt{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            acceptOn: ServerA,
        },
    }
    for _, fixture := range fixtures2 {
        s.T().Run(fixture.name, func(t *testing.T) {
            // 2.attach sk_lookup/redir_port into netns, 注意 reuseport bpf 程序的 attach 是直接 attach socketFd
            l, err := link.AttachNetNs(int(s.netns.Fd()), s.getBpfProg(fixture.progName))
            assert.NoError(s.T(), err)
            defer l.Close()

            serverFd, err := makeServer(fixture.socketType, nil, fixture.listenAt.ip, fixture.listenAt.port)
            if err != nil {
                logrus.Errorf("makeServer err: %v", err)
                return
            }
            defer unix.Close(serverFd)

            connectedFd := makeClient(fixture.socketType, fixture.connectTo.ip, fixture.connectTo.port)
            defer unix.Close(connectedFd)

            /* Put a connected socket in redirect map */
            err = s.objs.bpfMaps.RedirMap.Put(ServerA, uint64(connectedFd))

            /* Try to redirect TCP SYN / UDP packet to a connected socket */
            clientFd := makeClient(fixture.socketType, fixture.connectTo.ip, fixture.connectTo.port)
            defer unix.Close(clientFd)

            err = unix.Send(clientFd, []byte(getEchoData(fixture.acceptOn)), 0)
            if err != nil {
                logrus.Errorf("unix.Send err: %v", err)
                return
            }
            cbuf := make([]byte, 1024)
            n, _, _, _, err := unix.Recvmsg(serverFd, cbuf, nil, 0)
            if err != nil {
                logrus.Errorf("unix.Recvmsg err: %v", err)
                return
            }
            cbuf = cbuf[:n]
            logrus.Infof("server unix.Recvmsg from client: %s", string(cbuf))
        })
    }

    fixtures3 := []struct {
        name              string
        description       string
        prog1             string
        prog2             string
        reuseportProgName string
        socketType        int
        acceptOn          int
        reuseportHasConns bool
        listenAt          ListenAt
        connectTo         ConnectTo
        expected1         int
        expected2         int
    }{
        // test_multi_prog_lookup(), 验证 attach/run multi prog by order
        {
            name:       "multi prog pass pass",
            prog1:      "multi_prog_pass1",
            prog2:      "multi_prog_pass2",
            socketType: unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            expected1: 1,
            expected2: 1,
        },
        {
            name:       "multi prog drop drop",
            prog1:      "multi_prog_drop1",
            prog2:      "multi_prog_drop2",
            socketType: unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            expected1: 1,
            expected2: 1,
        },
        {
            name:       "multi prog pass drop",
            prog1:      "multi_prog_pass1",
            prog2:      "multi_prog_drop2",
            socketType: unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            expected1: 1,
            expected2: 1,
        },
        {
            name:       "multi prog drop pass",
            prog1:      "multi_prog_drop1",
            prog2:      "multi_prog_pass2",
            socketType: unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            expected1: 1,
            expected2: 1,
        },
        {
            name:       "multi prog pass redir",
            prog1:      "multi_prog_pass1",
            prog2:      "multi_prog_redir2",
            socketType: unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   INT_IP4,
                port: INT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            expected1: 1,
            expected2: 1,
        },
        {
            name:       "multi prog redir pass",
            prog1:      "multi_prog_redir1",
            prog2:      "multi_prog_pass2",
            socketType: unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   INT_IP4,
                port: INT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            expected1: 1,
            expected2: 1,
        },
        {
            name:       "multi prog drop redir",
            prog1:      "multi_prog_drop1",
            prog2:      "multi_prog_redir2",
            socketType: unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   INT_IP4,
                port: INT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            expected1: 1,
            expected2: 1,
        },
        {
            name:       "multi prog redir drop",
            prog1:      "multi_prog_redir1",
            prog2:      "multi_prog_drop2",
            socketType: unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   INT_IP4,
                port: INT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            expected1: 1,
            expected2: 1,
        },
        {
            name:       "multi prog redir redir",
            prog1:      "multi_prog_redir1",
            prog2:      "multi_prog_redir2",
            socketType: unix.SOCK_STREAM,
            listenAt: ListenAt{
                ip:   INT_IP4,
                port: INT_PORT,
            },
            connectTo: ConnectTo{
                ip:   EXT_IP4,
                port: EXT_PORT,
            },
            expected1: 1,
            expected2: 1,
        },
    }
    for _, fixture := range fixtures3 {
        s.T().Run(fixture.name, func(t *testing.T) {
            var err error
            key := Prog1
            done := 0
            err = s.objs.bpfMaps.RunMap.Put(unsafe.Pointer(&key), unsafe.Pointer(&done))
            if err != nil {
                logrus.Errorf("RunMap.Put err: %v", err)
                return
            }
            key = Prog2
            err = s.objs.bpfMaps.RunMap.Put(unsafe.Pointer(&key), unsafe.Pointer(&done))
            if err != nil {
                logrus.Errorf("RunMap.Put err: %v", err)
                return
            }

            // 2.attach sk_lookup/redir_port into netns, 注意 reuseport bpf 程序的 attach 是直接 attach socketFd
            l1, err := link.AttachNetNs(int(s.netns.Fd()), s.getBpfProg(fixture.prog1))
            assert.NoError(s.T(), err)
            defer l1.Close()
            l2, err := link.AttachNetNs(int(s.netns.Fd()), s.getBpfProg(fixture.prog2))
            assert.NoError(s.T(), err)
            defer l2.Close()

            serverFd, err := makeServer(fixture.socketType, nil, fixture.listenAt.ip, fixture.listenAt.port)
            if err != nil {
                logrus.Errorf("makeServer err: %v", err)
                return
            }
            defer unix.Close(serverFd)
            err = s.objs.bpfMaps.RedirMap.Put(uint32(ServerA), uint64(serverFd))
            if err != nil {
                logrus.Errorf("RunMap.Put err: %v", err)
                return
            }

            connectedFd := makeClient(fixture.socketType, fixture.connectTo.ip, fixture.connectTo.port)
            defer unix.Close(connectedFd)

            done = 0
            key = Prog1
            err = s.objs.bpfMaps.RunMap.Lookup(unsafe.Pointer(&key), unsafe.Pointer(&done))
            if err != nil {
                logrus.Errorf("RunMap.Lookup err: %v", err)
                return
            }
            if done == fixture.expected1 {
                logrus.Infof("RunMap.Lookup Prog1 successfully: expected %d, actual %d", fixture.expected1, done)
            } else {
                logrus.Errorf("RunMap.Lookup Prog1 error: expected %d, actual %d", fixture.expected1, done)
            }

            done = 0
            key = Prog2
            err = s.objs.bpfMaps.RunMap.Lookup(unsafe.Pointer(&key), unsafe.Pointer(&done))
            if err != nil {
                logrus.Errorf("RunMap.Lookup err: %v", err)
                return
            }
            if done == fixture.expected2 {
                logrus.Infof("RunMap.Lookup Prog2 successfully: expected %d, actual %d", fixture.expected2, done)
            } else {
                logrus.Errorf("RunMap.Lookup Prog2 error: expected %d, actual %d", fixture.expected2, done)
            }
        })
    }
}

func (s *SocketLookupSuite) getBpfProg(progName string) *ebpf.Program {
    switch progName {
    case "redir_port":
        return s.objs.bpfPrograms.RedirPort
    case "redir_ip4":
        return s.objs.bpfPrograms.RedirIp
    case "select_sock_a":
        return s.objs.bpfPrograms.SelectSockA
    case "select_sock_b":
        return s.objs.bpfPrograms.SelectSockB
    case "select_sock_a_no_reuseport":
        return s.objs.bpfPrograms.SelectSockA_noReuseport
    case "lookup_drop":
        return s.objs.bpfPrograms.LookupDrop
    case "reuseport_drop":
        return s.objs.bpfPrograms.ReuseportDrop
    case "sk_assign_eexist":
        return s.objs.bpfPrograms.SkAssignEexist
    case "sk_assign_replace_flag":
        return s.objs.bpfPrograms.SkAssignReplaceFlag
    case "sk_assign_null":
        return s.objs.bpfPrograms.SkAssignNull
    case "access_ctx_sk":
        return s.objs.bpfPrograms.AccessCtxSk
    case "ctx_narrow_access":
        return s.objs.bpfPrograms.CtxNarrowAccess
    case "sk_assign_estabsocknosupport":
        return s.objs.bpfPrograms.SkAssignEstabsocknosupport
    case "multi_prog_pass1":
        return s.objs.bpfPrograms.MultiProgPass1
    case "multi_prog_pass2":
        return s.objs.bpfPrograms.MultiProgPass2
    case "multi_prog_drop1":
        return s.objs.bpfPrograms.MultiProgDrop1
    case "multi_prog_drop2":
        return s.objs.bpfPrograms.MultiProgDrop2
    case "multi_prog_redir1":
        return s.objs.bpfPrograms.MultiProgRedir1
    case "multi_prog_redir2":
        return s.objs.bpfPrograms.MultiProgRedir2
    default:
        return nil
    }
}
