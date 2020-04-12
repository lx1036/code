package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

// ./network-namespace.sh
func main() {
	cmd := exec.Command("sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | // UTS Namespace, 主要隔离 nodename 和 domainname 两个系统标识。`hostname` + `hostname -b test`
			syscall.CLONE_NEWIPC | // IPC Namespace，隔离 IPC。`ipcs -q` + `ipcmk -Q`
			syscall.CLONE_NEWPID | // PID Namespace，隔离进程的 PID。`echo $$` `pstree -pl`
			syscall.CLONE_NEWNS | // Mount Namespace，隔离各个进程看到的文件系统挂载点视图。`ls /proc` `ps -ef`
			syscall.CLONE_NEWUSER | // User Namespace，隔离用户的用户组 ID`。 `id`
			syscall.CLONE_NEWNET, // Network Namespace，隔离网络设备、IP地址端口、单独协议栈以及 iptables/ipvs 规则。
		// `ip link list`，只有一个本地回环设备 lo，没有 veth 虚拟网卡，docker0 网桥
	}
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
