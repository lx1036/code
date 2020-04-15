package receivers

import "strings"

type ReceiverFactory struct {
}

func (factory *ReceiverFactory) BuildAll(receiverStr string)  {
	receivers := strings.Split(receiverStr, ",")
	receiver := receivers[0]

}
