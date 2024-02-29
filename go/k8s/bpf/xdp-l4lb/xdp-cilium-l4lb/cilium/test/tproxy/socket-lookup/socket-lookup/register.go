package main

import (
    "errors"
    "fmt"
    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "golang.org/x/sys/unix"
    "inet.af/netaddr"
    "os"
    "syscall"
)

var (
    pid           int
    netns         string
    registerLabel string
    protocol      string
    ip            string
    port          int
)

func init() {
    rootCmd.AddCommand(registerCmd)

    flags := registerCmd.PersistentFlags()
    flags.IntVarP(&pid, "pid", "", 0, "pid")
    viper.BindPFlag("pid", flags.Lookup("pid"))
    flags.StringVarP(&netns, "netns", "", "/proc/self/ns/net", "netns")
    viper.BindPFlag("netns", flags.Lookup("netns"))
    flags.StringVarP(&registerLabel, "label", "", "foo", "label")
    viper.BindPFlag("label", flags.Lookup("label"))
    flags.StringVarP(&protocol, "protocol", "", "tcp", "protocol")
    viper.BindPFlag("protocol", flags.Lookup("protocol"))
    flags.StringVarP(&ip, "ip", "", "127.0.0.1", "ip")
    viper.BindPFlag("ip", flags.Lookup("ip"))
    flags.IntVarP(&port, "port", "", 0, "port")
    viper.BindPFlag("port", flags.Lookup("port"))
}

// sk-lookup register 12345 foo tcp 127.0.0.1 80
var registerCmd = &cobra.Command{
    Use: "register",
    Run: func(cmd *cobra.Command, args []string) {
        if err := namespacesEqual(netns, fmt.Sprintf("/proc/%d/ns/net", pid)); err != nil {
            logr.Errorf("%v", err)
            return
        }

        registerPid()
    },
}

func registerPid() error {
    netaddrIP, err := netaddr.ParseIP(ip)
    if err != nil {
        return err
    }

    filter := []Predicate{
        IgnoreENOTSOCK(InetListener(protocol)),
        LocalAddress(netaddrIP, int(port)),
        FirstReuseport(),
    }

    // 进程 pid 打开 tcp://127.0.0.1:80 的所有 socket_fd
    files, err := Files(int(pid), filter...)
    if err != nil {
        return fmt.Errorf("pid %d: %w", pid, err)
    }

    defer func() {
        for _, f := range files {
            f.Close()
        }
    }()

    if err := registerFiles(registerLabel, files); err != nil {
        return fmt.Errorf("pid %d: %w", pid, err)
    }

    return nil
}

func registerFiles(label string, files []*os.File) error {
    if len(files) == 0 {
        return fmt.Errorf("no sockets")
    }

    dispatcher, err := CreateDispatcher()
    if err != nil {
        return err
    }
    defer dispatcher.Close()

    registered := make(map[Destinations]bool)
    for _, file := range files {
        dst, created, err := dispatcher.RegisterSocket(label, file)
        if err != nil {
            return fmt.Errorf("register fd: %w", err)
        }

        if registered[*dst] {
            return fmt.Errorf("found multiple sockets for destination %s", dst)
        }
        registered[*dst] = true

        cookie, _ := socketCookie(file)
        var msg string
        if created {
            msg = fmt.Sprintf("created destination %s", dst.String())
        } else {
            msg = fmt.Sprintf("updated destination %s", dst.String())
        }
        logrus.Infof("registered socket %s: %s", cookie, msg)
    }

    return nil
}

func socketCookie(conn syscall.Conn) (string, error) {
    var cookie uint64
    err := Control(conn, func(fd int) (err error) {
        cookie, err = unix.GetsockoptUint64(fd, unix.SOL_SOCKET, unix.SO_COOKIE)
        return
    })
    if err != nil {
        return "", fmt.Errorf("getsockopt(SO_COOKIE): %v", err)
    }

    return fmt.Sprintf("sk:%d", cookie), nil
}

// 比较两个 netns 是否相等，可以参考
func namespacesEqual(want, have string) error {
    var stat unix.Stat_t
    if err := unix.Stat(want, &stat); err != nil {
        return err
    }
    wantIno := stat.Ino

    if err := unix.Stat(have, &stat); err != nil {
        return err
    }
    haveIno := stat.Ino

    if wantIno != haveIno {
        return errors.New("can't register sockets from different network namespace")
    }

    return nil
}
