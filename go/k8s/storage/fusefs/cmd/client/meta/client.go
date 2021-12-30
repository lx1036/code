package meta

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
	"k8s.io/klog/v2"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/btree"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	RefreshMetaPartitionsInterval = time.Minute * 5

	DefaultBTreeDegree = 32
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
	statusConflictExtents
)

type MetaClient struct {
	sync.RWMutex

	volumeName   string
	masterAddrs  []string
	masterLeader string
	epoch        uint64

	totalSize  uint64
	usedSize   uint64
	inodeCount uint64
	status     VolStatus

	// Partition tree indexed by Start, in order to find a partition in which
	// a specific inode locate.
	inodesTree *btree.BTree

	rwPartitions []*Partition
	partitions   map[uint64]*Partition
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

func (metaClient *MetaClient) IsVolumeReadOnly() bool {
	return metaClient.status == ReadOnlyVol
}

func (metaClient *MetaClient) CreateInodeAndDentry(parentID fuseops.InodeID, filename string, mode, uid, gid uint32,
	target []byte) (*proto.InodeInfo, error) {
	parentPartition := metaClient.getPartitionByInode(parentID)
	if parentPartition == nil {
		return nil, fmt.Errorf(fmt.Sprintf("[CreateInodeAndDentry]fail to get parent partition id:%+v", parentID))
	}
	if len(parentPartition.LeaderAddr) == 0 {
		return nil, fmt.Errorf(fmt.Sprintf("partitionID %d has no leader address", parentPartition.PartitionID))
	}

	rwPartitions := metaClient.getRWPartitions()
	epoch := atomic.AddUint64(&metaClient.epoch, 1)
	length := len(rwPartitions)
	for i := 0; i < length; i++ {
		rwPartition := rwPartitions[(int(epoch)+i)%length] // ???
		if len(rwPartition.LeaderAddr) == 0 {
			klog.Infof(fmt.Sprintf("partitionID %d has no leader address", rwPartition.PartitionID))
			continue
		}
		status, inodeInfo, err := metaClient.createInode(rwPartition, mode, uid, gid, target)
		if err == nil && status == statusOK {
			// create dentry
			status, err = metaClient.createDentry(parentPartition, parentID, filename, inodeInfo.Inode, mode)
			if err == nil && (status == statusOK || status == statusExist) {
				return inodeInfo, nil
			} else {
				metaClient.unlinkInode()
				metaClient.evictInode()
				break
			}
		}
	}

	return nil, fmt.Errorf("fail to create inode/dentry for parentID:%d, filename:%s", parentID, filename)
}

func (metaClient *MetaClient) GetInode(parentID fuseops.InodeID) (*proto.InodeInfo, error) {
	parentPartition := metaClient.getPartitionByInode(parentID)
	if parentPartition == nil {
		return nil, fmt.Errorf(fmt.Sprintf("[CreateInodeAndDentry]fail to get parent partition id:%+v", parentID))
	}
	if len(parentPartition.LeaderAddr) == 0 {
		return nil, fmt.Errorf(fmt.Sprintf("partitionID %d has no leader address", parentPartition.PartitionID))
	}

	packet := proto.NewPacketReqID()
	packet.Opcode = proto.OpMetaCreateInode
	packet.PartitionID = parentPartition.PartitionID
	err := packet.MarshalData(&proto.InodeGetRequest{
		VolName:     metaClient.volumeName,
		PartitionID: parentPartition.PartitionID,
		Inode:       uint64(parentID),
	})
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", parentPartition.LeaderAddr)
	if err != nil {
		return nil, err
	}
	err = packet.WriteToConn(conn)
	if err != nil {
		return nil, err
	}
	if err = packet.ReadFromConn(conn, proto.ReadDeadlineTime); err != nil {
		return nil, err
	}
	if packet.ResultCode != proto.OpOk {
		return nil, fmt.Errorf("[GetInode]fail to get inode")
	}

	resp := new(proto.InodeGetResponse)
	err = json.Unmarshal(packet.Data, resp)
	if err != nil {
		return nil, err
	}

	return resp.Info, nil
}

func (metaClient *MetaClient) createInode(partition *Partition, mode, uid, gid uint32, target []byte) (int, *proto.InodeInfo, error) {
	packet := proto.NewPacketReqID()
	packet.Opcode = proto.OpMetaCreateInode
	packet.PartitionID = partition.PartitionID
	err := packet.MarshalData(&proto.CreateInodeRequest{
		VolName:     metaClient.volumeName,
		PartitionID: partition.PartitionID,
		Mode:        mode,
		Uid:         uid,
		Gid:         gid,
		Target:      target,
	})
	if err != nil {
		return statusUnknown, nil, err
	}

	conn, err := net.Dial("tcp", partition.LeaderAddr)
	if err != nil {
		return statusUnknown, nil, err
	}
	err = packet.WriteToConn(conn)
	if err != nil {
		return statusUnknown, nil, err
	}
	if err = packet.ReadFromConn(conn, proto.ReadDeadlineTime); err != nil {
		return statusUnknown, nil, err
	}
	if packet.ResultCode != proto.OpOk {
		return statusUnknown, nil, fmt.Errorf("[createInode]fail to create inode")
	}

	resp := new(proto.CreateInodeResponse)
	err = json.Unmarshal(packet.Data, resp)
	if err != nil {
		return statusUnknown, nil, err
	}

	return statusOK, resp.Info, nil
}

func (metaClient *MetaClient) createDentry(partition *Partition, parentID fuseops.InodeID, name string, inode uint64,
	mode uint32) (int, error) {
	if uint64(parentID) == inode {
		return statusExist, nil
	}

	packet := proto.NewPacketReqID()
	packet.Opcode = proto.OpMetaCreateDentry
	packet.PartitionID = partition.PartitionID
	err := packet.MarshalData(&proto.CreateDentryRequest{
		VolName:     metaClient.volumeName,
		PartitionID: partition.PartitionID,
		ParentID:    uint64(parentID),
		Inode:       inode,
		Name:        name,
		Mode:        mode,
	})
	if err != nil {
		return statusUnknown, err
	}

	conn, err := net.Dial("tcp", partition.LeaderAddr)
	if err != nil {
		return statusUnknown, err
	}
	err = packet.WriteToConn(conn)
	if err != nil {
		return statusUnknown, err
	}
	if err = packet.ReadFromConn(conn, proto.ReadDeadlineTime); err != nil {
		return statusUnknown, err
	}
	if packet.ResultCode != proto.OpOk {
		return statusUnknown, fmt.Errorf("[createDentry]fail to create inode")
	}

	return statusOK, nil
}

func (metaClient *MetaClient) getRWPartitions() []*Partition {
	if len(metaClient.rwPartitions) != 0 {
		return metaClient.rwPartitions
	}

	var partitions []*Partition
	for _, partition := range metaClient.partitions {
		partitions = append(partitions, partition)
	}

	return partitions
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
