package fuse

import (
	"unsafe"

	"k8s-lx1036/k8s/storage/fuse/internal/buffer"
)

////////////////////////////////////////////////////////////////////////
// buffer.InMessage
////////////////////////////////////////////////////////////////////////

// LOCKS_EXCLUDED(c.mu)
func (c *Connection) getInMessage() *buffer.InMessage {
	c.mu.Lock()
	x := (*buffer.InMessage)(c.inMessages.Get())
	c.mu.Unlock()

	if x == nil {
		x = new(buffer.InMessage)
	}

	return x
}

// LOCKS_EXCLUDED(c.mu)
func (c *Connection) putInMessage(x *buffer.InMessage) {
	c.mu.Lock()
	c.inMessages.Put(unsafe.Pointer(x))
	c.mu.Unlock()
}

////////////////////////////////////////////////////////////////////////
// buffer.OutMessage
////////////////////////////////////////////////////////////////////////

// LOCKS_EXCLUDED(c.mu)
func (c *Connection) getOutMessage() *buffer.OutMessage {
	c.mu.Lock()
	x := (*buffer.OutMessage)(c.outMessages.Get())
	c.mu.Unlock()

	if x == nil {
		x = new(buffer.OutMessage)
	}
	x.Reset()

	return x
}

// LOCKS_EXCLUDED(c.mu)
func (c *Connection) putOutMessage(x *buffer.OutMessage) {
	c.mu.Lock()
	c.outMessages.Put(unsafe.Pointer(x))
	c.mu.Unlock()
}
