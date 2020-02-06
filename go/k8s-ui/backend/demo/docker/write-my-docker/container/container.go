package container

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
)

func RunContainerInitProcess() error {
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {

	}

	setupMount()

	return nil
}

func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	defer pipe.Close()

	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Errorf("init read pipe %v", err)
		return nil
	}

	return strings.Split(string(msg), " ")
}

func setupMount() {
	pwd, err := os.Getwd()
	if err != nil {

	}

	pivotRoot(pwd)

	//mount proc
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
}

func pivotRoot(root string) error {

}
