package receivers

import (
	"k8s-lx1036/k8s-ui/backend/kubernetes/plugins/event/k8s-event-monitor/receivers/dingtalk"
	"strings"
)

type ReceiverFactory struct {
}

type Receiver interface {
}

func NewReceiverFactory() *ReceiverFactory {
	return &ReceiverFactory{}
}

func (factory *ReceiverFactory) BuildAll(receiverStr string) Receiver {
	receivers := strings.Split(receiverStr, ",")
	receiver := receivers[0]
	dingTalkReceiver := dingtalk.NewDingTalkReceiver(receiver)

	return dingTalkReceiver
}
