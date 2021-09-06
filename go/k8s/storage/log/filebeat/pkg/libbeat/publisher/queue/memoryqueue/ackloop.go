package memoryqueue

type ackLoop struct {
	broker *Broker
	sig    chan batchAckMsg
	lst    chanList

	totalACK   uint64
	totalSched uint64

	batchesSched uint64
	batchesACKed uint64

	processACK func(chanList, int)
}

func newACKLoop(b *Broker, processACK func(chanList, int)) *ackLoop {
	l := &ackLoop{broker: b}
	l.processACK = processACK
	return l
}
