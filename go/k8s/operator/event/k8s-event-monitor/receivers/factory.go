package receivers

import (
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
