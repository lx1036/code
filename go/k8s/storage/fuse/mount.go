package fuse

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/klog/v2"
)

// Server is an interface for any type that knows how to serve ops read from a
// connection.
type Server interface {
	// Read and serve ops from the supplied connection until EOF. Do not return
	// until all operations have been responded to. Must not be called more than
	// once.
	ServeOps(*Connection)
}

// Mount attempts to mount a file system on the given directory, using the
// supplied Server to serve connection requests. It blocks until the file
// system is successfully mounted.
func Mount(
	dir string,
	server Server,
	config *MountConfig) (*MountedFileSystem, error) {
	// Sanity check: make sure the mount point exists and is a directory. This
	// saves us from some confusing errors later on OS X.
	fi, err := os.Stat(dir)
	switch {
	case os.IsNotExist(err):
		return nil, err

	case err != nil:
		return nil, fmt.Errorf("Statting mount point: %v", err)

	case !fi.IsDir():
		return nil, fmt.Errorf("Mount point %s is not a directory", dir)
	}

	// Initialize the struct.
	mfs := &MountedFileSystem{
		dir:                 dir,
		joinStatusAvailable: make(chan struct{}),
	}

	// Begin the mounting process, which will continue in the background.
	ready := make(chan error, 1)
	dev, err := mount(dir, config, ready)
	if err != nil {
		return nil, fmt.Errorf("mount: %v", err)
	}

	// Choose a parent context for ops.
	cfgCopy := *config
	if cfgCopy.OpContext == nil {
		cfgCopy.OpContext = context.Background()
	}

	// Create a Connection object wrapping the device.
	connection, err := newConnection(
		cfgCopy,
		config.DebugLogger,
		config.ErrorLogger,
		dev)
	if err != nil {
		return nil, fmt.Errorf("newConnection: %v", err)
	}

	// Serve the connection in the background. When done, set the join status.
	go func() {
		server.ServeOps(connection)
		mfs.joinStatus = connection.close()
		close(mfs.joinStatusAvailable)
	}()

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigC
		close(mfs.joinStatusAvailable)
		err = unmount(mfs.Dir()) // TODO: mac 上会貌似 umount 失败
		if err != nil {
			klog.Errorf(fmt.Sprintf("umount %s err %+v", mfs.dir, err))
		}
		klog.Infof("Killed due to a received signal (%v)\n", sig)
		os.Exit(1)
	}()

	// Wait for the mount process to complete.
	if err := <-ready; err != nil {
		return nil, fmt.Errorf("mount (background): %v", err)
	}

	return mfs, nil
}
