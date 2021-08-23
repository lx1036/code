package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"k8s-lx1036/k8s/storage/fuse"
	"k8s-lx1036/k8s/storage/fuse/fuseutil"
	"k8s-lx1036/k8s/storage/fusefs/cmd/client/fs"

	"k8s.io/klog/v2"
)

const (
	MaxReadAhead     = 512 * 1024
	WriteBufPoolSize = 5 * 1024 * 1024 * 1024
)

const (
	// Mandatory
	MountPoint = "mountPoint"
	VolName    = "volName"
	Owner      = "owner"
	MasterAddr = "masterAddr"
	// Optional
	LogDir             = "logDir"
	LogLevel           = "logLevel"
	ProfPort           = "profPort"
	IcacheTimeout      = "icacheTimeout"
	LookupValid        = "lookupValid"
	AttrValid          = "attrValid"
	ReadRate           = "readRate"
	WriteRate          = "writeRate"
	EnSyncWrite        = "enSyncWrite"
	Rdonly             = "rdonly"
	WriteCache         = "writecache"
	KeepCache          = "keepcache"
	FullPathName       = "FullPathName"
	BufSize            = "bufSize"
	MaxMultiParts      = "maxMultiParts"
	MaxCacheInode      = "maxCacheInode"
	ReadDirBurst       = "readDirBurst"
	ReadDirLimit       = "readDirLimit"
	S3ObjectNameVerify = "s3ObjectNameVerify"
)

var (
	configFile = flag.String("c", "", "config file path")
)

// INFO: https://chubaofs.readthedocs.io/zh_CN/latest/design/client.html
// go run . -c ./fuse_360.json
func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	// mount config
	if len(*configFile) == 0 {
		klog.Fatalf(fmt.Sprintf("config file should not be empty"))
	}
	content, err := ioutil.ReadFile(*configFile)
	if err != nil {
		klog.Fatalf(fmt.Sprintf("read file err %v", err))
	}
	var mountOption fs.MountOption
	err = json.Unmarshal(content, &mountOption)
	if err != nil {
		klog.Fatalf(fmt.Sprintf("json unmarshal config file err %v", err))
	}

	fuseFS, err := fs.NewFuseFS(&mountOption)
	if err != nil {
		klog.Error(err)
		os.Exit(1)
	}

	registerInterceptedSignal(fuseFS)

	// mount filesystem
	server := fuseutil.NewFileSystemServer(fuseFS)
	mountConfig := &fuse.MountConfig{
		FSName:                  "sunfs-" + mountOption.Volname,
		Subtype:                 "sunfs", // `cat /proc/mounts | grep sunfs` -> xxx fuse.sunfs xxx
		ReadOnly:                mountOption.ReadOnly,
		DisableWritebackCaching: true,
	}

	mfs, err := fuse.Mount(mountOption.MountPoint, server, mountConfig)
	if err != nil {
		fuseFS.Destroy()
		klog.Error(err)
		os.Exit(1)
	}

	if err = mfs.Join(context.Background()); err != nil {
		klog.Errorf("mfs Joint returns error: %v", err)
		os.Exit(1)
	}
}

func registerInterceptedSignal(fuseFS *fs.FuseFS) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigC
		fuseFS.Destroy()
		klog.Infof("Killed due to a received signal (%v)\n", sig)
		os.Exit(1)
	}()
}
