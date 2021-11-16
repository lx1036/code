package main

import (
	"context"
	"crypto/sha512"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/containernetworking/cni/libcni"
	"github.com/containernetworking/cni/pkg/types"
	"k8s.io/klog/v2"
)

const (
	EnvCapabilityArgs = "CAP_ARGS"
	EnvCNIArgs        = "CNI_ARGS"
	EnvCNIIfname      = "CNI_IFNAME"

	DefaultNetDir = "/etc/cni/net.d"
	DefaultBINDir = "/usr/bin"

	CmdAdd   = "add"
	CmdCheck = "check"
	CmdDel   = "del"
)

func parseArgs(args string) ([][2]string, error) {
	var result [][2]string

	pairs := strings.Split(args, ";")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 || kv[0] == "" || kv[1] == "" {
			return nil, fmt.Errorf("invalid CNI_ARGS pair %q", pair)
		}

		result = append(result, [2]string{kv[0], kv[1]})
	}

	return result, nil
}

// INFO: @see https://github.com/containernetworking/cni/blob/master/cnitool/README.md
// go run . --cmd=add --pid=2049 --name=bandwidth --bin=./bin --conf=.
func main() {
	cmd := flag.String("cmd", CmdAdd, "cni cmd, e.g. add/check/del")
	pid := flag.Int("pid", 0, "container pid, e.g. /proc/2949/ns/net, 2949=`docker inspect ${container_id} | grep Pid`")
	name := flag.String("name", "cilium", "cni conf name, e.g. cilium in 05-cilium.conf 'name' field")
	cniBinDir := flag.String("bin", DefaultBINDir, "cni bin dir, e.g. /usr/bin. cilium-cni in 05-cilium.conf 'type' field")
	cniConfDir := flag.String("conf", DefaultNetDir, "cni net conf dir, e.g. /etc/cni/net.d")

	flag.Parse()

	if *pid == 0 {
		klog.Fatalf(fmt.Sprintf("pid is required"))
	}
	netns := fmt.Sprintf("/proc/%d/ns/net", *pid)

	netConfDir, _ := filepath.Abs(*cniConfDir)
	netconf, err := libcni.LoadConfList(netConfDir, *name)
	if err != nil {
		klog.Fatal(err)
	}

	var capabilityArgs map[string]interface{}
	capabilityArgsValue := os.Getenv(EnvCapabilityArgs)
	if len(capabilityArgsValue) > 0 {
		if err = json.Unmarshal([]byte(capabilityArgsValue), &capabilityArgs); err != nil {
			klog.Fatal(err)
		}
	}
	var cniArgs [][2]string
	args := os.Getenv(EnvCNIArgs)
	if len(args) > 0 {
		cniArgs, err = parseArgs(args)
		if err != nil {
			klog.Fatal(err)
		}
	}
	ifName, ok := os.LookupEnv(EnvCNIIfname)
	if !ok {
		ifName = "eth0"
	}

	// Generate the containerID by hashing the netns path
	s := sha512.Sum512([]byte(netns))
	containerID := fmt.Sprintf("cnitool-%x", s[:10])

	netBinDir, _ := filepath.Abs(*cniBinDir)
	cninet := libcni.NewCNIConfig(filepath.SplitList(netBinDir), nil)

	runtimeConf := &libcni.RuntimeConf{
		ContainerID:    containerID,
		NetNS:          netns,  // /proc/2949/ns/net
		IfName:         ifName, // eth0
		Args:           cniArgs,
		CapabilityArgs: capabilityArgs,
	}

	switch *cmd {
	case CmdAdd:
		var result types.Result
		result, err = cninet.AddNetworkList(context.TODO(), netconf, runtimeConf)
		if result != nil {
			_ = result.Print()
		}
	case CmdCheck:
		err = cninet.CheckNetworkList(context.TODO(), netconf, runtimeConf)
	case CmdDel:
		err = cninet.DelNetworkList(context.TODO(), netconf, runtimeConf)
	}

	if err != nil {
		klog.Fatal(err)
	}
}
