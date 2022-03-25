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
	RefreshMetaPartitionsInterval = time.Second * 60

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

type JsonResponse struct {
	Code int32           `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type Partition struct {
	PartitionID uint64
	Start       fuseops.InodeID // 起始 inodeID
	End         fuseops.InodeID // 终止 inodeID
	Members     []string
	LeaderAddr  string
	Status      int8
}

func (partition *Partition) Less(than btree.Item) bool {
	return partition.Start < than.(*Partition).Start
}

// TODO: packet TCP 这块可以使用 grpc pb 来标准化, @see go/k8s/storage/raft/hashicorp/raft/transport_tcp_test.go

// INFO: 调用 master-cluster API:
//  /admin/getVol: 获取 S3 endpoint
//  /client/volStat: 获取该 volume 的 totalSize/usedSize
//  /client/vol: 获取该 volume 分配的 meta partition 数据(包含 inode 范围，以及 partition LeaderAddress)

type MetaClient struct {
	sync.RWMutex

	volumeName   string
	volumeID     uint64
	masterAddrs  []string
	masterLeader string
	epoch        uint64
	S3Endpoint   string

	totalSize  uint64
	usedSize   uint64
	inodeCount uint64
	status     VolStatus

	// Partition tree indexed by Start, in order to find a partition in which
	// a specific inode locate.
	partitionsTree *btree.BTree
	rwPartitions   []*Partition
	partitions     map[uint64]*Partition
}

func NewMetaClient(volumeName, masterAddrs string) (*MetaClient, error) {
	addrs := strings.Split(masterAddrs, ",") // 127.0.0.1:9500,127.0.0.1:9600,127.0.0.1:9700
	metaClient := &MetaClient{
		volumeName:   volumeName,
		masterAddrs:  addrs,
		masterLeader: addrs[0], // TODO: 暂时选择第一个作为 leader address

		partitionsTree: btree.New(DefaultBTreeDegree),
		partitions:     make(map[uint64]*Partition),
	}

	if err := metaClient.getS3Endpoint(); err != nil {
		return nil, err
	}
	if err := metaClient.getVolStatInfo(); err != nil {
		return nil, err
	}
	if err := metaClient.getPartitionsForVol(); err != nil {
		return nil, err
	}

	go metaClient.start()

	return metaClient, nil
}

func (metaClient *MetaClient) start() {
	wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		metaClient.getVolStatInfo()
		metaClient.getPartitionsForVol()
		//metaClient.updateClientInfo()
	}, RefreshMetaPartitionsInterval)
}

type SimpleVolView struct {
	ID            uint64
	Name          string
	Owner         string
	MpReplicaNum  uint8
	Status        uint8
	Capacity      uint64 // GB
	MpCnt         int
	S3Endpoint    string
	BucketDeleted bool
}

/*
{
  "code": 0,
  "msg": "success",
  "data": {
    "ID": 78,
    "Name": "pvc-30e5e4f1-e04f-4d53-9c76-50a6bc35abb0",
    "Owner": "fusefs",
    "MpReplicaNum": 3,
    "Status": 0,
    "Capacity": 100,
    "MpCnt": 3,
    "S3Endpoint": "http://fusefs.s3.cn",
    "BucketDeleted": false
  }
}
*/
func (metaClient *MetaClient) getS3Endpoint() error {
	url := fmt.Sprintf("http://%s%s?name=%s", metaClient.masterLeader, "/admin/getVol", metaClient.volumeName)
	resp, err := http.Get(url)
	if err != nil {
		klog.Error(fmt.Sprintf("[getVolumeInfo]%v", err))
		return err
	}
	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		klog.Error(fmt.Sprintf("[getVolumeInfo]%v", err))
		return err
	}
	var jsonResponse JsonResponse
	if err = json.Unmarshal(data, &jsonResponse); err != nil {
		klog.Error(fmt.Sprintf("[getVolumeInfo]%v", err))
		return err
	}
	var simpleVolView SimpleVolView
	if err = json.Unmarshal(jsonResponse.Data, &simpleVolView); err != nil {
		klog.Error(fmt.Sprintf("[getVolumeInfo]%v", err))
		return err
	}

	metaClient.S3Endpoint = simpleVolView.S3Endpoint // "http://fusefs.s3.cn"
	return nil
}

/*
pv: pvc-30e5e4f1-e04f-4d53-9c76-50a6bc35abb0 大小为 100G
{
  "code": 0,
  "msg": "success",
  "data": {
    "Name": "pvc-30e5e4f1-e04f-4d53-9c76-50a6bc35abb0",
    "TotalSize": 107374182400,
    "UsedSize": 452,
    "Status": 0
  }
}
*/
func (metaClient *MetaClient) getVolStatInfo() error {
	url := fmt.Sprintf("http://%s%s?name=%s", metaClient.masterLeader, "/client/volStat", metaClient.volumeName)
	resp, err := http.Get(url)
	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		klog.Error(fmt.Sprintf("[updateVolStatInfo]%v", err))
		return err
	}
	var jsonResponse JsonResponse
	if err = json.Unmarshal(data, &jsonResponse); err != nil {
		klog.Error(fmt.Sprintf("[updateVolStatInfo]%v", err))
		return err
	}
	volStatInfo := &VolStatInfo{}
	if err = json.Unmarshal(jsonResponse.Data, volStatInfo); err != nil {
		klog.Error(fmt.Sprintf("[updateVolStatInfo]%v", err))
		return err
	}

	metaClient.totalSize = volStatInfo.TotalSize
	metaClient.usedSize = volStatInfo.UsedSize
	//metaClient.inodeCount = volStatInfo.InodeCount
	metaClient.status = volStatInfo.Status

	klog.Infof(fmt.Sprintf("[updateVolStatInfo]volName:%s, totalSize:%d, usedSize:%d, status:%d",
		metaClient.volumeName, metaClient.totalSize, metaClient.usedSize, metaClient.status))
	return nil
}

type Volume struct {
	Name           string
	Status         uint8
	MetaPartitions []*Partition
}

// INFO: get master `/vol` api, 获取该 volume 数据分布在哪些 partitions
/*
# meta cluster 有5台机器：
100.160.161.13:9021,100.160.161.14:9021,100.160.161.31:9021,100.160.161.49:9021,100.160.161.50:9021
# 默认选择 defaultReplicaNum=3 台机器作为 meta partition，且
* 第一台 meta partition 的 inode 范围: 0~16777216(defaultMetaPartitionInodeIDStep = 1 << 24 // 16MB)
* 第二台 meta partition 的 inode 范围：16777217~33554433(16777217+defaultMetaPartitionInodeIDStep)
* 第三台 meta partition 的 inode 范围：33554433~9223372036854775807(defaultMaxMetaPartitionInodeID = 1<<63 - 1)
{
    "code": 0,
    "msg": "success",
    "data": {
        "Name": "pvc-30e5e4f1-e04f-4d53-9c76-50a6bc35abb0",
        "Status": 0,
        "MetaPartitions": [
            {
                "PartitionID": 266,
                "Start": 0,
                "End": 16777216,
                "Members": [
                    "100.160.161.49:9021",
                    "100.160.161.13:9021",
                    "100.160.161.50:9021"
                ],
                "LeaderAddr": "100.160.161.50:9021",
                "Status": 2
            },
            {
                "PartitionID": 267,
                "Start": 16777217,
                "End": 33554433,
                "Members": [
                    "100.160.161.32:9021",
                    "100.160.161.31:9021",
                    "100.160.161.14:9021"
                ],
                "LeaderAddr": "100.160.161.32:9021",
                "Status": 2
            },
            {
                "PartitionID": 268,
                "Start": 33554434,
                "End": 9223372036854775807,
                "Members": [
                    "100.160.161.49:9021",
                    "100.160.161.13:9021",
                    "100.160.161.50:9021"
                ],
                "LeaderAddr": "100.160.161.50:9021",
                "Status": 2
            }
        ]
    }
}
*/
func (metaClient *MetaClient) getPartitionsForVol() error {
	url := fmt.Sprintf("http://%s%s?name=%s", metaClient.masterLeader, "/client/vol", metaClient.volumeName)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	var jsonResponse JsonResponse
	if err = json.Unmarshal(data, &jsonResponse); err != nil {
		klog.Error(fmt.Sprintf("[updateMetaPartitions]%v", err))
		return err
	}
	volume := &Volume{}
	if err = json.Unmarshal(jsonResponse.Data, volume); err != nil {
		klog.Error(fmt.Sprintf("[updateMetaPartitions]%v", err))
		return err
	}

	for _, partition := range volume.MetaPartitions {
		metaClient.replaceOrInsert(partition)
		if partition.Status == proto.ReadWrite {
			metaClient.rwPartitions = append(metaClient.rwPartitions, partition)
		}
	}

	return nil
}

func (metaClient *MetaClient) replaceOrInsert(partition *Partition) {
	value, ok := metaClient.partitions[partition.PartitionID]
	if ok {
		delete(metaClient.partitions, value.PartitionID)
		metaClient.partitionsTree.Delete(value)
	}

	metaClient.partitions[partition.PartitionID] = partition
	metaClient.partitionsTree.ReplaceOrInsert(partition)
}

func (metaClient *MetaClient) IsVolumeReadOnly() bool {
	return metaClient.status == ReadOnlyVol
}

// CreateInodeAndDentry INFO: 根据策略选择一个 rwPartition，然后从当前这个 partition 分配一个 inodeID
func (metaClient *MetaClient) CreateInodeAndDentry(parentID fuseops.InodeID, filename string, mode, uid, gid uint32,
	target []byte) (*proto.InodeInfo, error) {
	parentPartition := metaClient.getPartitionByInode(parentID)
	if parentPartition == nil {
		return nil, fmt.Errorf(fmt.Sprintf("[CreateInodeAndDentry]fail to get parent partition id:%+v", parentID))
	}
	if len(parentPartition.LeaderAddr) == 0 {
		return nil, fmt.Errorf(fmt.Sprintf("[CreateInodeAndDentry]partitionID %d has no leader address", parentPartition.PartitionID))
	}

	rwPartitions := metaClient.getRWPartitions()
	epoch := atomic.AddUint64(&metaClient.epoch, 1)
	length := len(rwPartitions)
	for i := 0; i < length; i++ {
		rwPartition := rwPartitions[(int(epoch)+i)%length] // ???
		if len(rwPartition.LeaderAddr) == 0 {
			klog.Infof(fmt.Sprintf("[CreateInodeAndDentry]partitionID %d has no leader address", rwPartition.PartitionID))
			continue
		}
		status, inodeInfo, err := metaClient.createInode(rwPartition, mode, uid, gid, target)
		if err == nil && status == statusOK {
			klog.Infof(fmt.Sprintf("[CreateInodeAndDentry]create inode:%d for filename:%s succefully", inodeInfo.Inode, filename))
			// create dentry
			status, err = metaClient.createDentry(parentPartition, parentID, filename, inodeInfo.Inode, mode)
			if err == nil && (status == statusOK || status == statusExist) {
				klog.Infof(fmt.Sprintf("[CreateInodeAndDentry]create dentry for filename:%s succefully", filename))
				return inodeInfo, nil
			} else {
				//metaClient.unlinkInode()
				//metaClient.evictInode()
				break
			}
		}
	}

	return nil, fmt.Errorf("[CreateInodeAndDentry]fail to create inode/dentry for parentID:%d, filename:%s", parentID, filename)
}

func (metaClient *MetaClient) GetInode(inodeID fuseops.InodeID) (*proto.InodeInfo, error) {
	parentPartition := metaClient.getPartitionByInode(inodeID)
	if parentPartition == nil {
		return nil, fmt.Errorf(fmt.Sprintf("[GetInode]fail to get parent partition id:%+v", inodeID))
	}
	if len(parentPartition.LeaderAddr) == 0 {
		return nil, fmt.Errorf(fmt.Sprintf("[GetInode]partitionID %d has no leader address", parentPartition.PartitionID))
	}

	packet := proto.NewPacketReqID()
	packet.Opcode = proto.OpMetaInodeGet
	err := packet.MarshalData(&proto.InodeGetRequest{
		VolName:     metaClient.volumeName,
		PartitionID: parentPartition.PartitionID,
		Inode:       uint64(inodeID),
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

	// {Inode:1 Mode:2147484159 Nlink:3 Size:0 Uid:0 Gid:0 Generation:1 ModifyTime:1639993844 CreateTime:1639993844 AccessTime:1639993844 Target:[] PInode:0}
	klog.Infof(fmt.Sprintf("[GetInode]InodeGetResponse:%+v", *(resp.Info)))
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

	klog.Infof(fmt.Sprintf("[createInode]CreateInodeResponse:%+v", *resp))
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

func (metaClient *MetaClient) LookupName(parentInodeID, currentInodeID fuseops.InodeID) (name string, err error) {

	return "", nil
}

// Lookup INFO: 根据 child name，从 parent inode 中查询出 inode
func (metaClient *MetaClient) Lookup(parentID fuseops.InodeID, name string) (inode uint64, mode uint32, err error) {
	parentPartition := metaClient.getPartitionByInode(parentID)
	if parentPartition == nil {
		return 0, 0, fmt.Errorf(fmt.Sprintf("[Lookup]fail to get parent partition id:%+v", parentID))
	}
	if len(parentPartition.LeaderAddr) == 0 {
		return 0, 0, fmt.Errorf(fmt.Sprintf("[Lookup]partitionID %d has no leader address", parentPartition.PartitionID))
	}

	status, inode, mode, err := metaClient.lookup(parentPartition, parentID, name)
	if err != nil || status != statusOK {
		return 0, 0, err
	}

	return inode, mode, nil
}

func (metaClient *MetaClient) lookup(partition *Partition, parentID fuseops.InodeID, name string) (status int, inode uint64, mode uint32, err error) {
	packet := proto.NewPacketReqID()
	packet.Opcode = proto.OpMetaLookup
	//packet.PartitionID = parentPartition.PartitionID
	if err = packet.MarshalData(&proto.LookupRequest{
		VolName:     metaClient.volumeName,
		PartitionID: partition.PartitionID,
		ParentID:    uint64(parentID),
		Name:        name,
	}); err != nil {
		return
	}

	conn, err := net.Dial("tcp", partition.LeaderAddr)
	if err != nil {
		return
	}
	if err = packet.WriteToConn(conn); err != nil {
		return
	}
	if err = packet.ReadFromConn(conn, proto.ReadDeadlineTime); err != nil {
		return
	}
	if packet.ResultCode != proto.OpOk {
		return 0, 0, 0, fmt.Errorf("[Lookup]fail to get inode for file/dir name:%s", name)
	}

	resp := new(proto.LookupResponse)
	if err = json.Unmarshal(packet.Data, resp); err != nil {
		return
	}

	klog.Infof(fmt.Sprintf("[Lookup]LookupResponse:%+v", *resp))
	return statusOK, resp.Inode, resp.Mode, nil
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
	metaClient.partitionsTree.DescendLessOrEqual(pivot, func(i btree.Item) bool { // DescendLessOrEqual???
		partition = i.(*Partition)
		if inodeID > partition.End || inodeID < partition.Start {
			partition = nil
		}
		return false
	})

	return partition
}

func (metaClient *MetaClient) ReadDir(inodeID fuseops.InodeID) ([]proto.Dentry, error) {
	partition := metaClient.getPartitionByInode(inodeID)
	if partition == nil {
		return nil, fmt.Errorf(fmt.Sprintf("[ReadDir]fail to get partition for inodeID:%+v", inodeID))
	}
	if len(partition.LeaderAddr) == 0 {
		return nil, fmt.Errorf(fmt.Sprintf("[ReadDir]partitionID %d has no leader address", partition.PartitionID))
	}

	packet := proto.NewPacketReqID()
	packet.Opcode = proto.OpMetaReadDir
	//packet.PartitionID = partition.PartitionID
	if err := packet.MarshalData(&proto.ReadDirRequest{
		VolName:     metaClient.volumeName,
		PartitionID: partition.PartitionID,
		ParentID:    uint64(inodeID),
	}); err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", partition.LeaderAddr)
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
		return nil, fmt.Errorf("[ReadDir]fail to get inode")
	}

	resp := new(proto.ReadDirResponse)
	err = json.Unmarshal(packet.Data, resp)
	if err != nil {
		return nil, err
	}

	klog.Infof(fmt.Sprintf("[ReadDir]ReadDirResponse:%+v", *resp))
	return resp.Children, nil
}

func (metaClient *MetaClient) SetAttr(inodeID fuseops.InodeID, valid, mode, uid, gid uint32, size, pino uint64) error {
	partition := metaClient.getPartitionByInode(inodeID)
	if partition == nil {
		return fmt.Errorf(fmt.Sprintf("[ReadDir]fail to get partition for inodeID:%+v", inodeID))
	}
	if len(partition.LeaderAddr) == 0 {
		return fmt.Errorf(fmt.Sprintf("[ReadDir]partitionID %d has no leader address", partition.PartitionID))
	}

	packet := proto.NewPacketReqID()
	packet.Opcode = proto.OpMetaSetattr
	//packet.PartitionID = partition.PartitionID
	if err := packet.MarshalData(&proto.SetAttrRequest{
		VolName:     metaClient.volumeName,
		PartitionID: partition.PartitionID,
		Inode:       uint64(inodeID),
		Mode:        mode,
		Uid:         uid,
		Gid:         gid,
		Size:        size,
		Pino:        pino,
		Valid:       valid,
	}); err != nil {
		return err
	}
	conn, err := net.Dial("tcp", partition.LeaderAddr)
	if err != nil {
		return err
	}
	err = packet.WriteToConn(conn)
	if err != nil {
		return err
	}
	if err = packet.ReadFromConn(conn, proto.ReadDeadlineTime); err != nil {
		return err
	}
	if packet.ResultCode != proto.OpOk {
		return fmt.Errorf("[ReadDir]fail to get inode")
	}

	return nil
}

func (metaClient *MetaClient) Statfs() (total, used, inodeCount uint64) {
	return metaClient.totalSize, metaClient.usedSize, metaClient.inodeCount
}
