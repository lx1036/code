package main

import (
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
    "github.com/containernetworking/plugins/pkg/ns"
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "golang.org/x/sys/unix"
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
            PinPath: pinPath, // pin 下 map
        },
    }
    err = loadBpfObjects(&objs, opts)
    assert.NoError(s.T(), err)
    s.objs = objs
}

func (s *SocketLookupSuite) TearDownSuite() {
    s.netns.Close()
    s.objs.Close()
}

func (s *SocketLookupSuite) SetupTest() {
}

func (s *SocketLookupSuite) TearDownTest() {
}

// CGO_ENABLED=0 go test -cover -v -coverprofile=coverage.out -testify.m ^TestSocketLookupWithReuseport$ .
func (s *SocketLookupSuite) TestSocketLookupWithReuseport() {
    fixtures := []struct {
        name              string
        progName          string
        reuseportProgName string
        socketType        int
        acceptOn          int
        reuseportHasConns bool
    }{
        {
            name:       "TCP IPv4 redir port",
            progName:   "redir_port",
            socketType: unix.SOCK_STREAM,
        },
        {
            name:       "TCP IPv4 redir addr",
            progName:   "redir_ip4",
            socketType: unix.SOCK_STREAM,
        },
        {
            // INFO: reuseport 这块还是没想明白???
            name:              "TCP IPv4 redir with reuseport",
            progName:          "select_sock_a",
            reuseportProgName: "select_sock_b",
            socketType:        unix.SOCK_STREAM,
            acceptOn:          ServerB, /* ServerA */ // 两个 socketFd 都挂载了 sk_reuseport/select_sock_b 程序(???)，如果写 ServerA，bof 程序必须查找 KEY_SERVER_A
        },
        {
            name:              "TCP IPv4 redir skip reuseport",
            progName:          "select_sock_a_no_reuseport",
            reuseportProgName: "select_sock_b",
            socketType:        unix.SOCK_STREAM,
            acceptOn:          ServerA,
        },

        // UDP 一直没调试成功，好像 bpf 程序对 udp 不起作用，尽管适用于 udp???
        //{
        //    name:       "UDP IPv4 redir port",
        //    progName:   "redir_port",
        //    socketType: unix.SOCK_DGRAM,
        //},
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

        {
            name:       "TCP IPv4 drop on lookup",
            progName:   "lookup_drop",
            socketType: unix.SOCK_STREAM,
        },
        //{
        //    name:       "UDP IPv4 drop on lookup",
        //    progName:   "lookup_drop",
        //    socketType: unix.SOCK_DGRAM,
        //},
        {
            name:              "TCP IPv4 drop on reuseport",
            progName:          "select_sock_a",
            reuseportProgName: "reuseport_drop",
            socketType:        unix.SOCK_STREAM,
        },
        {
            name:              "UDP IPv4 drop on reuseport",
            progName:          "select_sock_a",
            reuseportProgName: "reuseport_drop",
            socketType:        unix.SOCK_DGRAM,
        },
    }
    for _, fixture := range fixtures {
        s.T().Run(fixture.name, func(t *testing.T) {
            // 2.attach sk_lookup/redir_port into netns, 注意 reuseport bpf 程序的 attach 是直接 attach socketFd
            l, err := link.AttachNetNs(int(s.netns.Fd()), s.getBpfProg(fixture.progName))
            assert.NoError(s.T(), err)
            defer l.Close()

            // 3.update map
            serverFds := makeServer(fixture.socketType, s.getBpfProg(fixture.reuseportProgName))
            for i := 0; i < MAX_SERVERS; i++ {
                key := uint32(i)
                value := uint64(serverFds[i])
                err = s.objs.bpfMaps.RedirMap.Put(key, value)
            }
            defer func() {
                for i := 0; i < MAX_SERVERS; i++ {
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
                serverFds2 := makeServer(fixture.socketType, s.getBpfProg(fixture.reuseportProgName))
                /* Connect the extra socket to itself */
                reuseConnFd := serverFds2[0]
                reuseConnSockAddr, err := unix.Getsockname(reuseConnFd)
                assert.NoError(s.T(), err)
                err = unix.Connect(reuseConnFd, reuseConnSockAddr)
                assert.NoError(s.T(), err)
            }

            // 4.tcp echo
            clientFd := makeClient(fixture.socketType)
            defer unix.Close(clientFd)
            tcpEcho(clientFd, serverFds[fixture.acceptOn])
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
