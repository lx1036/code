package receivers

import (
)

type ReceiverManager struct {
}

func NewReceiverManager(receiver Receiver) *ReceiverManager {

	return &ReceiverManager{}
}

func (manager *ReceiverManager) Send(events common.Events) {

}
