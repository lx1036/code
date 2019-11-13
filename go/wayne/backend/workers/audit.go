package workers

import (
	"k8s-lx1036/wayne/backend/bus"
	"sync"
)

const QueueAudit = "audit"

var (
	lock          sync.Mutex
	queueDeclared = false
)

type AuditWorker struct {
	*BaseMessageWorker
}

func NewAuditWorker(b *bus.Bus) (*AuditWorker, error) {
	baseWorker := NewBaseMessageWorker(b, QueueAudit)
	w := &AuditWorker{baseWorker}
	w.BaseMessageWorker.MessageWorker = w

	lock.Lock()
	defer lock.Unlock()

	if !queueDeclared {
		if _, err := w.Bus.Channel.QueueDeclare(QueueAudit, true, false, false, false, nil); err != nil {
			return nil, err
		}
		if err := w.Bus.Channel.QueueBind(QueueAudit, bus.RoutingKeyRequest, w.Bus.Name, false, nil); err != nil {
			return nil, err
		}
		w.Bus.Channel.Qos(1, 0, false)
		queueDeclared = true
	}

	return w, nil
}
