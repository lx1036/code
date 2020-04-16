package receivers

import (
	"k8s-lx1036/k8s-ui/backend/kubernetes/plugins/event/k8s-event-monitor/common"
)

type ReceiverManager struct {

}


func NewReceiverManager(receiver Receiver) *ReceiverManager {

	return &ReceiverManager{}
}

func (manager *ReceiverManager) Send(events common.Events) {

}
