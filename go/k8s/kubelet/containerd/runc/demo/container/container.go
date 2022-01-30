package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
)

// INFO: 创建一个 container
// go run . run sh
// 要在 linux 上运行测试
func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	}
}

func run() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | // UTS(Unix Timesharing System) Namespace, 主要隔离 nodename 和 domainname 两个系统标识。`hostname` + `hostname -b test`
			syscall.CLONE_NEWIPC | // IPC Namespace(进程通信)，隔离 IPC。`ipcs -q` + `ipcmk -Q`
			syscall.CLONE_NEWPID | // PID Namespace(进程隔离)，隔离进程的 PID。`echo $$` `pstree -pl`
			syscall.CLONE_NEWNS | // Mount Namespace(Filesystem文件系统)，隔离各个进程看到的文件系统挂载点视图。`ls /proc` `ps -ef`
			syscall.CLONE_NEWUSER | // User Namespace(用户系统)，隔离用户的用户组 ID`。 `id`
			syscall.CLONE_NEWNET, // Network Namespace(网络)，隔离网络设备、IP地址端口、单独协议栈以及 iptables/ipvs 规则。验证：`ip link list`
		// `ip link list`，只有一个本地回环设备 lo，没有 veth 虚拟网卡，docker0 网桥
	}
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func child() {
	fmt.Printf("running %v as PID %d\n", os.Args[2:], os.Getpid())
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
