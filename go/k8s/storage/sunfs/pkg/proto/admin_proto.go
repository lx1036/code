package proto

// api
const (
	// Admin APIs
	AdminGetCluster           = "/admin/getCluster"
	AdminGetVolMountClient    = "/admin/getVolMountClient"
	AdminDeleteVol            = "/vol/delete"
	AdminUpdateVol            = "/vol/update"
	AdminCreateVol            = "/admin/createVol"
	AdminGetVol               = "/admin/getVol"
	AdminClusterFreeze        = "/cluster/freeze"
	AdminGetIP                = "/admin/getIp"
	AdminCreateMP             = "/metaPartition/create"
	AdminSetMetaNodeThreshold = "/threshold/set"

	// Client APIs
	ClientVol            = "/client/vol"
	ClientMetaPartition  = "/client/metaPartition"
	ClientVolStat        = "/client/volStat"
	ClientMetaPartitions = "/client/metaPartitions"
	ClientVolMount       = "/client/volMount"
	ClientVolUnMount     = "/client/volUnMount"
	ClientVolMountUpdate = "/client/volMountUpdate"

	//raft node APIs
	AddRaftNode    = "/raftNode/add"
	RemoveRaftNode = "/raftNode/remove"

	// Node APIs
	AddMetaNode                    = "/metaNode/add"
	DecommissionMetaNode           = "/metaNode/decommission"
	GetMetaNode                    = "/metaNode/get"
	AdminLoadMetaPartition         = "/metaPartition/load"
	AdminDecommissionMetaPartition = "/metaPartition/decommission"
	AdminAddMetaReplica            = "/metaReplica/add"
	AdminDeleteMetaReplica         = "/metaReplica/delete"

	// Operation response
	GetMetaNodeTaskResponse = "/metaNode/response" // Method: 'POST', ContentType: 'application/json'

	GetTopologyView = "/topo/get"
)

// The following defines the status of a disk or a partition.
const (
	ReadOnly    = 1
	ReadWrite   = 2
	Unavailable = -1
)

type ClusterInfo struct {
	Cluster string
	Ip      string
}

type ClientInfo struct {
	Id         uint64
	Ip         string
	Hostname   string
	Version    string
	MemoryUsed string
	MountVol   string
	MountPoint string
	System     string
	Expiration string
}

// SimpleVolView defines the simple view of a volume
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

// HTTPReply uniform response structure
type HTTPReply struct {
	Code int32       `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}
