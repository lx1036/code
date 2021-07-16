package meta

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"k8s-lx1036/k8s/storage/sunfs/pkg/util"

	"github.com/google/btree"
)

const (
	HostsSeparator                = ","
	RefreshMetaPartitionsInterval = time.Minute * 5
)

const (
	MaxMountRetryLimit = 5
	MountRetryInterval = time.Second * 5
)

const (
	statusUnknown int = iota
	statusOK
	statusExist
	statusNoent
	statusFull
	statusAgain
	statusError
	statusInval
	statusNotPerm
)

type MetaWrapper struct {
	sync.RWMutex
	cluster    string
	localIP    string
	volname    string
	owner      string
	S3Endpoint string
	master     util.MasterHelper
	conns      *util.ConnectPool

	// Partitions and ranges should be modified together. So do not
	// use partitions and ranges directly. Use the helper functions instead.

	// Partition map indexed by ID
	partitions map[uint64]*MetaPartition

	// Partition tree indexed by Start, in order to find a partition in which
	// a specific inode locate.
	ranges *btree.BTree

	rwPartitions []*MetaPartition
	epoch        uint64

	totalSize uint64
	usedSize  uint64

	clientId uint64
}

func (mw *MetaWrapper) Statfs() (total, used uint64) {
	mw.updateVolStatInfo()
	total = atomic.LoadUint64(&mw.totalSize)
	used = atomic.LoadUint64(&mw.usedSize)
	return
}

func (mw *MetaWrapper) refresh() {
	wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		err := mw.updateMetaPartitions()
		if err != nil {
			klog.Warningf("[refresh]%v", err)
		}
		err = mw.updateVolStatInfo()
		if err != nil {
			klog.Warningf("[refresh]%v", err)
		}
		//mw.updateClientInfo()
	}, RefreshMetaPartitionsInterval)
}

func NewMetaWrapper(volname, owner, masterHosts string) (*MetaWrapper, error) {
	mw := new(MetaWrapper)
	mw.volname = volname
	mw.owner = owner
	master := strings.Split(masterHosts, HostsSeparator)
	mw.master = util.NewMasterHelper()
	for _, ip := range master {
		mw.master.AddNode(ip)
	}
	mw.conns = util.NewConnectPool()
	mw.partitions = make(map[uint64]*MetaPartition)
	mw.ranges = btree.New(32)
	mw.rwPartitions = make([]*MetaPartition, 0)
	err := mw.updateClusterInfo()
	if err != nil {
		return nil, err
	}
	err = mw.updateVolStatInfo()
	if err != nil {
		return nil, err
	}
	err = mw.updateVolSimpleInfo()

	if err := mw.updateMetaPartitions(); err != nil {
		return nil, fmt.Errorf("init meta wrapper failed err: %v", err)
	}

	// TODO: refresh
	//go mw.refresh()

	return mw, nil
}
