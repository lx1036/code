package buffer

// The maximum read size that we expect to ever see from the kernel, used for
// calculating the size of out messages.
//
// Experimentally determined on OS X.
const MaxReadSize = 1 << 20
