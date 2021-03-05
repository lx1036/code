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
