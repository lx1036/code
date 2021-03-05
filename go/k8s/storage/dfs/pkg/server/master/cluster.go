package master

import (
	"k8s-lx1036/k8s/storage/dfs/pkg/raftstore"
	"sync"
)

// Cluster stores all the cluster-level information.
type Cluster struct {
	Name                string
	vols                map[string]*Vol
	volMountClients     map[string]*MountClients
	buckets             map[string]*DeleteBucketInfo
	metaNodes           sync.Map
	volMutex            sync.RWMutex // volume mutex
	volMountClientMutex sync.RWMutex // volume mount client mutex
	bucketMutex         sync.RWMutex
	createVolMutex      sync.RWMutex // create volume mutex
	mnMutex             sync.RWMutex // meta node mutex
	leaderInfo          *LeaderInfo
	cfg                 *clusterConfig
	retainLogs          uint64
	idAlloc             *IDAllocator
	t                   *topology
	metaNodeStatInfo    *nodeStatInfo
	volStatInfo         sync.Map
	DisableAutoAllocate bool
	fsm                 *MetadataFsm
	partition           raftstore.Partition
}

func (c *Cluster) scheduleTask() {
	c.scheduleToCheckHeartbeat()
	c.scheduleToCheckMetaPartitions()
	c.scheduleToUpdateStatInfo()
	c.scheduleToCheckVolStatus()
	c.scheduleToLoadMetaPartitions()
	c.scheduleToCheckVolMountClients()
}

func newCluster(name string, leaderInfo *LeaderInfo, fsm *MetadataFsm,
	partition raftstore.Partition, cfg *clusterConfig) (c *Cluster) {
	c = new(Cluster)
	c.Name = name
	c.leaderInfo = leaderInfo
	c.vols = make(map[string]*Vol, 0)
	c.volMountClients = make(map[string]*MountClients, 0)
	c.buckets = make(map[string]*DeleteBucketInfo, 0)
	c.cfg = cfg
	c.t = newTopology()
	c.metaNodeStatInfo = new(nodeStatInfo)
	c.fsm = fsm
	c.partition = partition
	c.idAlloc = newIDAllocator(c.fsm.store, c.partition)
	return
}
