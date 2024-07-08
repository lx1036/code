package agent

import (
	"encoding/gob"
	"net"
	"sync"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/monitor/agent/listener"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/monitor/payload"
)

// listenerv1_2 implements the cilium-node-monitor API protocol compatible with
// cilium 1.2
// cleanupFn is called on exit
type listenerv1_2 struct {
	// Used to prevent queue from getting closed multiple times.
	once sync.Once

	conn  net.Conn
	queue chan *payload.Payload

	cleanupFn func(listener.MonitorListener)
}

func (ml *listenerv1_2) Enqueue(pl *payload.Payload) {
	select {
	case ml.queue <- pl:
	default:
		log.Debug("Per listener queue is full, dropping message")
	}
}

func (ml *listenerv1_2) Version() listener.Version {
	return listener.Version1_2
}

// Close closes the underlying socket and payload queue.
func (ml *listenerv1_2) Close() {
	ml.once.Do(func() {
		ml.conn.Close()
		close(ml.queue)
	})
}

// drainQueue encodes and sends monitor payloads to the listener. It is
// intended to be a goroutine.
func (ml *listenerv1_2) drainQueue() {
	defer func() {
		ml.cleanupFn(ml)
	}()

	// 相当于给 accept conn 写 payload 数据，而 payload 从 queue 里获取
	enc := gob.NewEncoder(ml.conn)
	for pl := range ml.queue {
		if err := pl.EncodeBinary(enc); err != nil {
			switch {
			case listener.IsDisconnected(err):
				log.Debug("Listener disconnected")
				return

			default:
				log.WithError(err).Warn("Removing listener due to write failure")
				return
			}
		}
	}
}

func newListenerv1_2(c net.Conn, queueSize int, cleanupFn func(listener.MonitorListener)) *listenerv1_2 {
	ml := &listenerv1_2{
		conn:      c,
		queue:     make(chan *payload.Payload, queueSize),
		cleanupFn: cleanupFn,
	}

	go ml.drainQueue()

	return ml
}
