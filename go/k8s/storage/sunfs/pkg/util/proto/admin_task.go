package proto

import (
	"fmt"
	"time"
)

const (
	TaskFailed       = 2
	TaskStart        = 0
	TaskSucceeds     = 1
	TaskRunning      = 3
	ResponseInterval = 5
	ResponseTimeOut  = 100
	MaxSendCount     = 5
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

// NewAdminTask returns a new adminTask.
func NewAdminTask(opCode uint8, opAddr string, request interface{}) *AdminTask {
	return &AdminTask{
		ID:           fmt.Sprintf("addr[%v]_op[%v]", opAddr, opCode),
		OpCode:       opCode,
		OperatorAddr: opAddr,
		CreateTime:   time.Now().Unix(),
		Request:      request,
	}
}
