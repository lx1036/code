package buffer

// The maximum read size that we expect to ever see from the kernel, used for
// calculating the size of out messages.
//
// For 4 KiB pages, this is 1024 KiB (cf. https://github.com/torvalds/linux/blob/15db16837a35d8007cb8563358787412213db25e/fs/fuse/fuse_i.h#L38-L40)
const MaxReadSize = 1 << 20
