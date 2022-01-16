package fuse

import (
	"bytes"
	"fmt"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"unsafe"
)

// INFO: 参考 mount_linux.go 优化下 mount() 函数。已经优化。
//  参考下 https://github.com/jacobsa/fuse/blob/master/mount_darwin.go#L85-L107 监听在 /dev/macfusexxx 文件上
//  **[基于Fuse的用户态文件系统性能优化几点建议](https://zhuanlan.zhihu.com/p/68085075)**

const MaxWriteSize = 1 << 16 // 64 KB

// Create a FUSE FS on the specified mount point. The returned
// mount point is always absolute.
func mount(mountPoint string, opts *MountConfig, ready chan<- error) (*os.File, error) {
	fd, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, os.NewSyscallError("socketpair", err.(syscall.Errno))
	}
	// Wrap the sockets into os.File objects that we will pass off to fusermount.
	writeFile := os.NewFile(uintptr(fd[0]), "fusermount-writes")
	defer writeFile.Close()
	readFile := os.NewFile(uintptr(fd[1]), "fusermount-reads")
	defer readFile.Close()

	bin, err := findFusermount()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(bin,
		"-o", opts.toOptionsString(),
		"-o", fmt.Sprintf("iosize=%s", strconv.FormatUint(MaxWriteSize, 10)),
		mountPoint)
	cmd.ExtraFiles = []*os.File{writeFile} // fd would be (index + 3)
	cmd.Env = append(os.Environ(),
		"_FUSE_CALL_BY_LIB=",
		"_FUSE_DAEMON_PATH="+os.Args[0],
		"_FUSE_COMMFD=3",
		"_FUSE_COMMVERS=2",
		"MOUNT_OSXFUSE_CALL_BY_LIB=",
		"MOUNT_OSXFUSE_DAEMON_PATH="+os.Args[0])

	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	klog.Infof(fmt.Sprintf("cmd: %s", cmd.String()))
	if err = cmd.Start(); err != nil {
		return nil, err
	}

	readFileFD, err := getConnection(readFile)
	if err != nil {
		return nil, err
	}

	go func() {
		// wait inside a goroutine or otherwise it would block forever for unknown reasons
		if err := cmd.Wait(); err != nil {
			err = fmt.Errorf("mount_osxfusefs failed: %v. Stderr: %s, Stdout: %s",
				err, errOut.String(), out.String())
			klog.Error(err)
		} else {
			klog.Infof(fmt.Sprintf("cmd start successfully!!!"))
		}

		ready <- err
		close(ready)
	}()

	// golang sets CLOEXEC on file descriptors when they are
	// acquired through normal operations (e.g. open).
	// Buf for fd, we have to set CLOEXEC manually
	syscall.CloseOnExec(readFileFD)

	// Turn the FD into an os.File
	return os.NewFile(uintptr(readFileFD), "fuse"), err
}

// get fd from file
func getConnection(local *os.File) (int, error) {
	var data [4]byte
	control := make([]byte, 4*256)

	// n, oobn, recvflags, from, errno  - todo: error checking.
	_, oobn, _, _, err := syscall.Recvmsg(int(local.Fd()), data[:], control[:], 0)
	if err != nil {
		return 0, err
	}

	message := *(*syscall.Cmsghdr)(unsafe.Pointer(&control[0]))
	fd := *(*int32)(unsafe.Pointer(uintptr(unsafe.Pointer(&control[0])) + syscall.SizeofCmsghdr))

	if message.Type != syscall.SCM_RIGHTS {
		return 0, fmt.Errorf("getConnection: recvmsg returned wrong control type: %d", message.Type)
	}
	if oobn <= syscall.SizeofCmsghdr {
		return 0, fmt.Errorf("getConnection: too short control message. Length: %d", oobn)
	}
	if fd < 0 {
		return 0, fmt.Errorf("getConnection: fd < 0: %d", fd)
	}
	return int(fd), nil
}

func findFusermount() (string, error) {
	binPaths := []string{
		"/Library/Filesystems/macfuse.fs/Contents/Resources/mount_macfuse", // INFO: 目前macfuse主要是这一个
		"/Library/Filesystems/osxfuse.fs/Contents/Resources/mount_osxfuse",
	}

	for _, path := range binPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no FUSE mount utility found")
}
