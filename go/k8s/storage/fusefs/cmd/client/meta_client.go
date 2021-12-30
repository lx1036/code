package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
	"k8s.io/klog/v2"
	"net/http"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/btree"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	RefreshMetaPartitionsInterval = time.Minute * 5

	DefaultBTreeDegree = 32
)

type MetaClient struct {
	sync.RWMutex

	volumeName   string
	masterAddrs  []string
	masterLeader string

	totalSize  uint64
	usedSize   uint64
	inodeCount uint64
	status     VolStatus

	// Partition tree indexed by Start, in order to find a partition in which
	// a specific inode locate.
	inodesTree *btree.BTree
}

func NewMetaClient(volumeName, owner, masterAddrs string) (*MetaClient, error) {
	addrs := strings.Split(masterAddrs, ",") // 127.0.0.1:9500,127.0.0.1:9600,127.0.0.1:9700
	metaClient := &MetaClient{
		volumeName:   volumeName,
		masterAddrs:  addrs,
		masterLeader: addrs[0], // TODO: 暂时选择第一个作为 leader address

		inodesTree: btree.New(DefaultBTreeDegree),
	}

	go metaClient.start()

	return metaClient, nil
}

func (metaClient *MetaClient) start() {
	wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		metaClient.updateMetaPartitions()
		metaClient.updateVolStatInfo()
		//metaClient.updateClientInfo()
	}, RefreshMetaPartitionsInterval)
}

func (metaClient *MetaClient) isVolumeReadOnly() bool {
	return metaClient.status == ReadOnlyVol
}

func (metaClient *MetaClient) CreateInodeAndDentry(parentID fuseops.InodeID) (*proto.InodeInfo, error) {
	partition := metaClient.getPartitionByInode(parentID)
	if partition == nil {
		klog.Errorf(fmt.Sprintf("[CreateInodeAndDentry]fail to get parent partition id:%+v", parentID))
		return nil, syscall.ENOENT
	}

}

func (metaClient *MetaClient) getPartitionByInode(inodeID fuseops.InodeID) *Partition {
	metaClient.Lock()
	defer metaClient.Unlock()

	var partition *Partition
	pivot := &Partition{Start: inodeID}
	metaClient.inodesTree.DescendLessOrEqual(pivot, func(i btree.Item) bool { // DescendLessOrEqual???
		partition = i.(*Partition)
		if inodeID > partition.End || inodeID < partition.Start {
			partition = nil
		}
		return false
	})

	return partition
}

type VolStatInfo struct {
	Name        string
	TotalSize   uint64
	UsedSize    uint64
	UsedRatio   string
	EnableToken bool
	InodeCount  uint64
	Status      VolStatus
}
type VolStatus uint8

const (
	ReadWriteVol   VolStatus = 1
	MarkDeletedVol VolStatus = 2
	ReadOnlyVol    VolStatus = 3
)

func (metaClient *MetaClient) updateVolStatInfo() error {
	url := fmt.Sprintf("%s?name=%s", metaClient.masterLeader, metaClient.volumeName)
	resp, err := http.Get(url)
	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		klog.Error(fmt.Sprintf("[updateVolStatInfo]%v", err))
		return err
	}

	volStatInfo := &VolStatInfo{}
	if err = json.Unmarshal(data, volStatInfo); err != nil {
		klog.Error(fmt.Sprintf("[updateVolStatInfo]%v", err))
		return err
	}

	metaClient.totalSize = volStatInfo.TotalSize
	metaClient.usedSize = volStatInfo.UsedSize
	metaClient.inodeCount = volStatInfo.InodeCount
	metaClient.status = volStatInfo.Status
	return nil
}

func (metaClient *MetaClient) Statfs() (total, used, inodeCount uint64) {
	return metaClient.totalSize, metaClient.usedSize, metaClient.inodeCount
}

type Partition struct {
	PartitionID uint64
	Start       fuseops.InodeID
	End         fuseops.InodeID
	Members     []string
	LeaderAddr  string
	Status      int8
}

func (partition *Partition) Less(than btree.Item) bool {
	return partition.Start < than.(*Partition).Start
}
