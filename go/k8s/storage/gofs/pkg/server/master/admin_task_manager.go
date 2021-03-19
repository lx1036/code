package master

import (
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/gofs/pkg/util"
	"k8s-lx1036/k8s/storage/gofs/pkg/util/proto"
)

//const
const (
	// the maximum number of tasks that can be handled each time
	MaxTaskNum         = 30
	TaskWorkerInterval = time.Second * time.Duration(2)
)

// AdminTaskManager sends administration commands to the metaNode.
type AdminTaskManager struct {
	clusterID  string
	targetAddr string
	TaskMap    map[string]*proto.AdminTask
	sync.RWMutex
	exitCh   chan struct{}
	connPool *util.ConnectPool
}
