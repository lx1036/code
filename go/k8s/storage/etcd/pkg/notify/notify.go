package notify

import (
	"sync"
)

// Notifier is a thread safe struct that can be used to send notification about
// some event to multiple consumers.
type Notifier struct {
	mu      sync.RWMutex
	channel chan struct{}
}

// NewNotifier returns new notifier
func NewNotifier() *Notifier {
	return &Notifier{
		channel: make(chan struct{}),
	}
}

// Receive returns channel that can be used to wait for notification.
// Consumers will be informed by closing the channel.
func (n *Notifier) Receive() <-chan struct{} {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.channel
}

// Notify closes the channel passed to consumers and creates new channel to used
// for next notification.
func (n *Notifier) Notify() {
	newChannel := make(chan struct{})
	n.mu.Lock()
	channelToClose := n.channel
	n.channel = newChannel
	n.mu.Unlock()
	close(channelToClose)
}
