package buffer

// The maximum fuse write request size that InMessage can acommodate.
//
// As of kernel 4.20 Linux accepts writes up to 256 pages or 1MiB
const MaxWriteSize = 1 << 20
