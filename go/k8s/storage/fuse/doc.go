// Package fuse enables writing and mounting user-space file systems.
//
// The primary elements of interest are:
//
//  *  The fuseops package, which defines the operations that fuse might send
//     to your userspace daemon.
//
//  *  The Server interface, which your daemon must implement.
//
//  *  fuseutil.NewFileSystemServer, which offers a convenient way to implement
//     the Server interface.
//
//  *  Mount, a function that allows for mounting a Server as a file system.
//
// Make sure to see the examples in the sub-packages of samples/, which double
// as tests for this package: http://godoc.org/k8s-lx1036/k8s/storage/fuse/samples
//
// In order to use this package to mount file systems on OS X, the system must
// have FUSE for OS X installed (see http://osxfuse.github.io/). Do note that
// there are several OS X-specific oddities; grep through the documentation for
// more info.
package fuse
