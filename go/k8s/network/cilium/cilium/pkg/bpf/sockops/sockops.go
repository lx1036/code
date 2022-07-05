package sockops

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/cgroup"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/defaults"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/option"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/datapath/loader"

	"k8s.io/klog/v2"
)

// INFO: bpftool 程序是 linux 内核源码自带的，代码在 https://github.com/lx1036/linux/blob/master/tools/bpf/bpftool/Makefile
//  借助这篇文章一起食用更佳：http://arthurchiao.art/blog/socket-acceleration-with-ebpf-zh/

const (
	contextTimeout = 5 * time.Minute

	sockMap = "cilium_sock_ops"

	eSockops = "bpf_sockops"
	eIPC     = "bpf_redir"

	cSockops = "bpf_sockops.c"
	oSockops = "bpf_sockops.o"

	cIPC = "bpf_redir.c"
	oIPC = "bpf_redir.o"
)

// SockmapEnable will compile sockops programs and attach the sockops programs
// to the cgroup. After this all TCP connect events will be filtered by a BPF
// sockops program.
func SockmapEnable() error {
	err := bpfCompileProg(cSockops, oSockops)
	if err != nil {
		klog.Error(err)
		return err
	}
	progID, mapID, err := bpfLoadAttachProg(oSockops, eSockops, sockMap)
	if err != nil {
		klog.Error(err)
		return err
	}
	klog.Infof("Sockmap Enabled: bpf_sockops prog_id %d and map_id %d loaded", progID, mapID)
	return nil
}

// SkmsgEnable will compile and attach the SK_MSG programs to the
// sockmap. After this all sockets added to the cilium_sock_ops will
// have sendmsg/sendfile calls running through BPF program.
func SkmsgEnable() error {
	err := bpfCompileProg(cIPC, oIPC)
	if err != nil {
		klog.Error(err)
		return err
	}

	err = bpfLoadMapProg(oIPC, eIPC)
	if err != nil {
		klog.Error(err)
		return err
	}
	klog.Info("Sockmsg Enabled, bpf_redir loaded")
	return nil
}

// SockmapDisable will detach any sockmap programs from cgroups then "unload"
// all the programs and maps associated with it. Here "unload" just means
// deleting the file associated with the map.
func SockmapDisable() {
	mapName := filepath.Join(defaults.DefaultMapPrefix, sockMap)
	bpftoolDetach(eSockops)
	bpftoolUnload(eSockops)
	bpftoolUnload(mapName)
}

// SkmsgDisable "unloads" the SK_MSG program. This simply deletes
// the file associated with the program.
func SkmsgDisable() {
	bpftoolUnload(eIPC)
}

// INFO: clang && llc 编译 bpf c 代码为字节码 bpf obj 文件
// #clang ... | llc ...
func bpfCompileProg(src string, dst string) error {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	srcpath := filepath.Join("sockops", src) // 文件在 bpf/sockops
	outpath := filepath.Join(dst)

	err := loader.Compile(ctx, srcpath, outpath)
	if err != nil {
		return fmt.Errorf("failed compile %s: %s", srcpath, err)
	}
	return nil
}

// 先加载 load，然后再 attach 到 sockmap
func bpfLoadAttachProg(object string, load string, mapName string) (int, int, error) {
	sockopsObj := filepath.Join(option.Config.StateDir, object)
	mapID := 0

	err := bpftoolLoad(sockopsObj, load)
	if err != nil {
		return 0, 0, err
	}
	err = bpftoolAttach(load) // `bpftool cgroup attach $cgrp sock_ops /sys/fs/bpf/$bpfObject`
	if err != nil {
		return 0, 0, err
	}

	if mapName != "" {
		mapID, err = bpftoolGetMapID(load, mapName)
		if err != nil {
			return 0, mapID, err
		}

		err = bpftoolPinMapID(mapName, mapID)
		if err != nil {
			return 0, mapID, err
		}
	}

	return 0, mapID, nil
}

func bpfLoadMapProg(object string, load string) error {
	sockopsObj := filepath.Join(option.Config.StateDir, object)
	err := bpftoolLoad(sockopsObj, load)
	if err != nil {
		return err
	}

	progID, err := bpftoolGetProgID(load)
	if err != nil {
		return err
	}
	_mapID, err := bpftoolGetMapID(eSockops, sockMap)
	mapID := strconv.Itoa(_mapID)
	if err != nil {
		return err
	}

	err = bpftoolMapAttach(progID, mapID)
	if err != nil {
		return err
	}
	return nil
}

// #rm $bpfObject
func bpftoolUnload(bpfObject string) {
	bpffs := filepath.Join(bpf.GetMapRoot(), bpfObject) // /sys/fs/bpf/tc/globals/xxx
	os.Remove(bpffs)
}

// #bpftool cgroup attach $cgrp sock_ops /sys/fs/bpf/$bpfObject
func bpftoolAttach(bpfObject string) error {
	prog := "bpftool"
	bpffs := filepath.Join(bpf.GetMapRoot(), bpfObject)
	cgrp := cgroup.GetCgroupRoot()
	args := []string{"cgroup", "attach", cgrp, "sock_ops", "pinned", bpffs}
	out, err := exec.Command(prog, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to attach %s: %s: %s", bpfObject, err, out)
	}
	return nil
}

// #bpftool cgroup detach $cgrp sock_ops /sys/fs/bpf/$bpfObject
func bpftoolDetach(bpfObject string) error {
	prog := "bpftool"
	bpffs := filepath.Join(bpf.GetMapRoot(), bpfObject)
	cgrp := cgroup.GetCgroupRoot()
	args := []string{"cgroup", "detach", cgrp, "sock_ops", "pinned", bpffs}
	out, err := exec.Command(prog, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to detach %s: %s: %s", bpfObject, err, out)
	}
	return nil
}

// #bpftool prog show pinned /sys/fs/bpf/bpf_sockops
// #bpftool map show id 21
func bpftoolGetMapID(progName string, mapName string) (int, error) {
	bpffs := filepath.Join(bpf.GetMapRoot(), progName)
	prog := "bpftool"
	args := []string{"prog", "show", "pinned", bpffs}
	output, err := exec.Command(prog, args...).CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("Failed to load %s: %s: %s", progName, err, output)
	}

	// Find the mapID out of the bpftool output
	s := strings.Fields(string(output))
	for i := range s {
		if s[i] == "map_ids" {
			id := strings.Split(s[i+1], ",")
			for j := range id {
				args := []string{"map", "show", "id", id[j]}
				output, err := exec.Command(prog, args...).CombinedOutput()
				if err != nil {
					return 0, err
				}

				if strings.Contains(string(output), mapName) {
					mapID, _ := strconv.Atoi(id[j])
					return mapID, nil
				}
			}
			break
		}
	}

	return 0, nil
}

// #bpftool map pin id map_id /sys/fs/bpf/tc/globals/xxx
func bpftoolPinMapID(mapName string, mapID int) error {
	mapFile := filepath.Join(bpf.GetMapRoot(), defaults.DefaultMapPrefix, mapName)
	prog := "bpftool"
	args := []string{"map", "pin", "id", strconv.Itoa(mapID), mapFile}
	out, err := exec.Command(prog, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to pin map %d(%s): %s: %s", mapID, mapName, err, out)
	}

	return nil
}

// #bpftool prog load $bpfObject /sys/fs/bpf/sockops
func bpftoolLoad(bpfObject string, bpfFsFile string) error {
	sockopsMaps := [...]string{
		"cilium_lxc",
		"cilium_ipcache",
		"cilium_metric",
		"cilium_events",
		"cilium_sock_ops",
		"cilium_ep_to_policy",
		"cilium_proxy4", "cilium_proxy6",
		"cilium_lb6_reverse_nat", "cilium_lb4_reverse_nat",
		"cilium_lb6_services", "cilium_lb4_services",
		"cilium_lb6_rr_seq", "cilium_lb4_seq",
		"cilium_lb6_rr_seq", "cilium_lb4_seq",
	}
	prog := "bpftool"
	var mapArgList []string
	bpffs := filepath.Join(bpf.GetMapRoot(), bpfFsFile)
	maps, err := ioutil.ReadDir(filepath.Join(bpf.GetMapRoot(), "/tc/globals/"))
	if err != nil {
		return err
	}

	for _, f := range maps {
		// Ignore all backing files
		if strings.HasPrefix(f.Name(), "..") {
			continue
		}

		use := func() bool {
			for _, n := range sockopsMaps {
				if f.Name() == n {
					return true
				}
			}
			return false
		}()

		if !use {
			continue
		}

		mapString := []string{"map", "name", f.Name(), "pinned", filepath.Join(bpf.GetMapRoot(), "/tc/globals/", f.Name())}
		mapArgList = append(mapArgList, mapString...)
	}

	// bpftool -m prog load bpf_redir.o /sys/fs/bpf/bpf_redir map name sock_ops_map pinned /sys/fs/bpf/sock_ops_map
	args := []string{"-m", "prog", "load", bpfObject, bpffs}
	args = append(args, mapArgList...)
	out, err := exec.Command(prog, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to load %s: %s: %s", bpfObject, err, out)
	}
	return nil
}

// BPF programs and sockmaps working on cgroups
// #bpftool prog attach progID msg_verdict id mapID
func bpftoolMapAttach(progID string, mapID string) error {
	prog := "bpftool"
	args := []string{"prog", "attach", "id", progID, "msg_verdict", "id", mapID}
	out, err := exec.Command(prog, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to attach prog(%s) to map(%s): %s: %s", progID, mapID, err, out)
	}
	return nil
}

// #bpftool prog show pinned /sys/fs/bpf/bpf_redir
func bpftoolGetProgID(progName string) (string, error) {
	bpffs := filepath.Join(bpf.GetMapRoot(), progName)
	prog := "bpftool"
	args := []string{"prog", "show", "pinned", bpffs}
	output, err := exec.Command(prog, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Failed to load %s: %s: %s", progName, err, output)
	}

	// Scrap the prog_id out of the bpftool output after libbpf is dual licensed
	// we will use programatic API.
	s := strings.Fields(string(output))
	if s[0] == "" {
		return "", fmt.Errorf("Failed to find prog %s: %s", progName, err)
	}
	progID := strings.Split(s[0], ":")
	return progID[0], nil
}
