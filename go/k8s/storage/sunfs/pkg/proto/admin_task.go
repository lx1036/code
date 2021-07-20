package proto

// AdminTask defines the administration task.
type AdminTask struct {
	ID           string
	PartitionID  uint64
	OpCode       uint8
	OperatorAddr string
	Status       int8
	SendTime     int64
	CreateTime   int64
	SendCount    uint8
	Request      interface{}
	Response     interface{}
}
