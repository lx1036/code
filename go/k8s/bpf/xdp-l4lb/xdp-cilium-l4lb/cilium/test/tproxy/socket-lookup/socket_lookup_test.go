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
)

/**
"TCP IPv4 redir port"
"TCP IPv4 redir addr"
"TCP IPv4 redir with reuseport"
"TCP IPv4 redir skip reuseport"

"UDP IPv4 redir port"
"UDP IPv4 redir addr"
"UDP IPv4 redir with reuseport"
"UDP IPv4 redir and reuseport with conns"
"UDP IPv4 redir skip reuseport"

*/

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
// test_redirect_lookup()
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

        // UDP 一直没调试成功，好像 bpf 程序对 udp 不起作用，尽管适用于 udp???
        {
            name:        "UDP IPv4 redir port",
            description: "127.0.0.1:7007 > 127.0.0.2:8008",
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
        //{
        //    name:       "UDP IPv4 redir addr",
        //    progName:   "redir_ip4",
        //    socketType: unix.SOCK_DGRAM,
        //},
        //{
        //    name:              "UDP IPv4 redir with reuseport",
        //    progName:          "select_sock_a",
        //    reuseportProgName: "select_sock_b",
        //    socketType:        unix.SOCK_DGRAM,
        //    acceptOn:          ServerB, /* ServerA */
        //},
        //{
        //    name:              "UDP IPv4 redir and reuseport with conns",
        //    progName:          "select_sock_a",
        //    reuseportProgName: "select_sock_b",
        //    socketType:        unix.SOCK_DGRAM,
        //    acceptOn:          ServerB, /* ServerA */
        //    reuseportHasConns: true,
        //},
        //{
        //    name:              "UDP IPv4 redir skip reuseport",
        //    progName:          "select_sock_a_no_reuseport",
        //    reuseportProgName: "select_sock_b",
        //    socketType:        unix.SOCK_DGRAM,
        //    acceptOn:          ServerA,
        //},

        //{
        //    name:       "TCP IPv4 drop on lookup",
        //    progName:   "lookup_drop",
        //    socketType: unix.SOCK_STREAM,
        //},
        ////{
        ////    name:       "UDP IPv4 drop on lookup",
        ////    progName:   "lookup_drop",
        ////    socketType: unix.SOCK_DGRAM,
        ////},
        //{
        //    name:              "TCP IPv4 drop on reuseport",
        //    progName:          "select_sock_a",
        //    reuseportProgName: "reuseport_drop",
        //    socketType:        unix.SOCK_STREAM,
        //},
        //{
        //    name:              "UDP IPv4 drop on reuseport",
        //    progName:          "select_sock_a",
        //    reuseportProgName: "reuseport_drop",
        //    socketType:        unix.SOCK_DGRAM,
        //},
    }
    for _, fixture := range fixtures {
        s.T().Run(fixture.name, func(t *testing.T) {
            // 2.attach sk_lookup/redir_port into netns, 注意 reuseport bpf 程序的 attach 是直接 attach socketFd
            l, err := link.AttachNetNs(int(s.netns.Fd()), s.getBpfProg(fixture.progName))
            assert.NoError(s.T(), err)
            defer l.Close()

            // 3.update map
            serverFds := makeServer(fixture.socketType, s.getBpfProg(fixture.reuseportProgName), fixture.listenAt.ip, fixture.listenAt.port)
            for i := 0; i < len(serverFds); i++ {
                key := uint32(i)
                value := uint64(serverFds[i])
                err = s.objs.bpfMaps.RedirMap.Put(key, value)
            }
            defer func() {
                for i := 0; i < len(serverFds); i++ {
                    unix.Close(serverFds[i])
                }
            }()

            /* Regular UDP socket lookup with reuseport behaves
             * differently when reuseport group contains connected
             * sockets. Check that adding a connected UDP socket to the
             * reuseport group does not affect how reuseport works with
             * BPF socket lookup.
             */
            if fixture.reuseportHasConns {
                /* Add an extra socket to reuseport group */
                serverFds2 := makeServer(fixture.socketType, s.getBpfProg(fixture.reuseportProgName), fixture.listenAt.ip, fixture.listenAt.port)
                defer func() {
                    for i := 0; i < len(serverFds2); i++ {
                        unix.Close(serverFds2[i])
                    }
                }()
                // INFO: Connect the extra socket to itself, 注意还可以 connect socket itself
                reuseConnFd := serverFds2[0]
                reuseConnSockAddr, err := unix.Getsockname(reuseConnFd)
                assert.NoError(s.T(), err)
                err = unix.Connect(reuseConnFd, reuseConnSockAddr)
                assert.NoError(s.T(), err)
            }

            // 4.tcp echo
            clientFd := makeClient(fixture.socketType, fixture.connectTo.ip, fixture.connectTo.port)
            defer unix.Close(clientFd)

            if fixture.socketType == unix.SOCK_STREAM {
                tcpEcho(clientFd, serverFds[fixture.acceptOn])
            } else {
                udpEcho(clientFd, serverFds[fixture.acceptOn])
            }
        })
    }
}

func (s *SocketLookupSuite) TestSocketAssign() {

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
    default:
        return nil
    }
}
