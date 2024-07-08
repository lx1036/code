package api

// ClusterInfo is a snapshot of cluster by cache.
type ClusterInfo struct {
	Jobs           map[JobID]*JobInfo
	Nodes          map[string]*NodeInfo
	Queues         map[QueueID]*QueueInfo
	NamespaceInfo  map[NamespaceName]*NamespaceInfo
	RevocableNodes map[string]*NodeInfo
	NodeList       []string
	CSINodesStatus map[string]*CSINodeStatusInfo
}
