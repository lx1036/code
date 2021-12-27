package proto

import (
	"fmt"
	"time"
)

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

func NewAdminTask(opCode uint8, opAddr string, request interface{}) (t *AdminTask) {
	return &AdminTask{
		ID:           fmt.Sprintf("addr[%v]_op[%v]", t.OperatorAddr, t.OpCode),
		OpCode:       opCode,
		OperatorAddr: opAddr,
		CreateTime:   time.Now().Unix(),
		Request:      request,
	}
}
