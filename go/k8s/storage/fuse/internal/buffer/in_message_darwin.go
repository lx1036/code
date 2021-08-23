package buffer

// The maximum fuse write request size that InMessage can acommodate.
//
// Experimentally, OS X appears to cap the size of writes to 1 MiB, regardless
// of whether a larger size is specified in the mount options.
const MaxWriteSize = 1 << 20
