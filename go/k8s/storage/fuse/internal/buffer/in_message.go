package buffer

import (
	"fmt"
	"io"
	"syscall"
	"unsafe"

	"k8s-lx1036/k8s/storage/fuse/internal/fusekernel"
)

// All requests read from the kernel, without data, are shorter than
// this.
var pageSize int

// We size the buffer to have enough room for a fuse request plus data
// associated with a write request.
var bufSize int

func init() {
	pageSize = syscall.Getpagesize()
	bufSize = pageSize + MaxWriteSize
}

// An incoming message from the kernel, including leading fusekernel.InHeader
// struct. Provides storage for messages and convenient access to their
// contents.
type InMessage struct {
	remaining []byte
	storage   []byte
}

// Initialize with the data read by a single call to r.Read. The first call to
// Consume will consume the bytes directly after the fusekernel.InHeader
// struct.
func (m *InMessage) Init(r io.Reader) error {
	m.storage = make([]byte, bufSize, bufSize)
	n, err := r.Read(m.storage[:])
	if err != nil {
		return err
	}

	// Make sure the message is long enough.
	const headerSize = unsafe.Sizeof(fusekernel.InHeader{})
	if uintptr(n) < headerSize {
		return fmt.Errorf("Unexpectedly read only %d bytes.", n)
	}

	m.remaining = m.storage[headerSize:n]

	// Check the header's length.
	if int(m.Header().Len) != n {
		return fmt.Errorf(
			"Header says %d bytes, but we read %d",
			m.Header().Len,
			n)
	}

	return nil
}

// Return a reference to the header read in the most recent call to Init.
func (m *InMessage) Header() *fusekernel.InHeader {
	return (*fusekernel.InHeader)(unsafe.Pointer(&m.storage[0]))
}

// Return the number of bytes left to consume.
func (m *InMessage) Len() uintptr {
	return uintptr(len(m.remaining))
}

// Consume the next n bytes from the message, returning a nil pointer if there
// are fewer than n bytes available.
func (m *InMessage) Consume(n uintptr) unsafe.Pointer {
	if m.Len() == 0 || n > m.Len() {
		return nil
	}

	p := unsafe.Pointer(&m.remaining[0])
	m.remaining = m.remaining[n:]

	return p
}

// Equivalent to Consume, except returns a slice of bytes. The result will be
// nil if Consume would fail.
func (m *InMessage) ConsumeBytes(n uintptr) []byte {
	if n > m.Len() {
		return nil
	}

	b := m.remaining[:n]
	m.remaining = m.remaining[n:]

	return b
}
