package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"k8s-lx1036/k8s/storage/gofs/pkg/client"
	"k8s-lx1036/k8s/storage/gofs/pkg/util/config"

	"github.com/jacobsa/fuse"
	"github.com/jacobsa/fuse/fuseutil"

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
	configFile       = flag.String("c", "", "config file path")
	configVersion    = flag.Bool("v", false, "show version")
	configForeground = flag.Bool("f", false, "run foreground")
)

func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	/*
	 * LoadConfigFile should be checked before start daemon, since it will
	 * call os.Exit() w/o notifying the parent process.
	 */
	cfg, err := config.LoadConfigFile(*configFile)
	if err != nil {
		klog.Error(err)
		os.Exit(1)
	}

	opt, err := parseMountOption(cfg)
	if err != nil {
		klog.Errorf("parseMountOption err: %v", err)
		os.Exit(1)
	}

	super, err := client.NewSuper(opt)
	if err != nil {
		klog.Error(err)
		os.Exit(1)
	}

	registerInterceptedSignal(super)

	// mount filesystem
	server := fuseutil.NewFileSystemServer(super)
	mntcfg := &fuse.MountConfig{
		FSName:                  "polefs-" + opt.Volname,
		Subtype:                 "polefs",
		ReadOnly:                opt.Rdonly,
		DisableWritebackCaching: true,
	}

	mfs, err := fuse.Mount(opt.MountPoint, server, mntcfg)
	if err != nil {
		super.Destroy()
		klog.Error(err)
		os.Exit(1)
	}

	if err = mfs.Join(context.Background()); err != nil {
		klog.Errorf("mfs Joint returns error: %v", err)
		os.Exit(1)
	}
}

func registerInterceptedSignal(super *client.Super) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigC
		super.Destroy()
		klog.Infof("Killed due to a received signal (%v)\n", sig)
		os.Exit(1)
	}()
}

func parseMountOption(cfg *config.Config) (*client.MountOption, error) {
	var err error
	opt := new(client.MountOption)
	opt.Config = cfg

	rawmnt := cfg.GetString(MountPoint)
	opt.MountPoint, err = filepath.Abs(rawmnt)
	if err != nil {
		return nil, fmt.Errorf("invalide mount point (%s) ", rawmnt)
	}

	opt.Volname = cfg.GetString(VolName)
	opt.Owner = cfg.GetString(Owner)
	opt.Master = cfg.GetString(MasterAddr)
	opt.Logpath = cfg.GetString(LogDir)
	opt.Loglvl = cfg.GetString(LogLevel)
	opt.Profport = cfg.GetString(ProfPort)
	opt.IcacheTimeout = parseConfigString(cfg, IcacheTimeout)
	opt.LookupValid = parseConfigString(cfg, LookupValid)
	opt.AttrValid = parseConfigString(cfg, AttrValid)
	opt.ReadRate = parseConfigString(cfg, ReadRate)
	opt.WriteRate = parseConfigString(cfg, WriteRate)
	opt.EnSyncWrite = parseConfigString(cfg, EnSyncWrite)
	opt.Rdonly = cfg.GetBool(Rdonly)
	opt.WriteCache = cfg.GetBool(WriteCache)
	opt.KeepCache = cfg.GetBool(KeepCache)
	opt.FullPathName = cfg.GetBoolWithDefault(FullPathName, true)
	opt.S3ObjectNameVerify = cfg.GetBoolWithDefault(S3ObjectNameVerify, true)
	opt.BufSize = cfg.GetInt64WithDefault(BufSize, WriteBufPoolSize)
	opt.MaxMultiParts = cfg.GetInt(MaxMultiParts)
	opt.MaxCacheInode = cfg.GetInt(MaxCacheInode)
	if opt.MaxMultiParts <= 0 {
		opt.MaxMultiParts = 60
	}
	opt.ReadDirBurst = cfg.GetInt(ReadDirBurst)
	opt.ReadDirLimit = cfg.GetInt(ReadDirLimit)

	if opt.MountPoint == "" || opt.Volname == "" || opt.Owner == "" || opt.Master == "" {
		return nil, fmt.Errorf("invalid config file: lack of mandatory fields, mountPoint(%s), volName(%s), owner(%s), masterAddr(%s)",
			opt.MountPoint, opt.Volname, opt.Owner, opt.Master)
	}

	opt.S3Cfg, err = config.ParseS3Config(cfg)

	return opt, nil
}

func parseConfigString(cfg *config.Config, keyword string) int64 {
	var ret int64 = -1
	rawstr := cfg.GetString(keyword)
	if rawstr != "" {
		val, err := strconv.Atoi(rawstr)
		if err == nil {
			ret = int64(val)
			fmt.Println(fmt.Sprintf("keyword[%v] value[%v]", keyword, ret))
		}
	}
	return ret
}
