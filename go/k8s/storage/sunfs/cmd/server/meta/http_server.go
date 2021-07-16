package meta

import "net/http"

// register the APIs
func (m *MetaNode) registerAPIHandler() (err error) {
	http.HandleFunc("/getPartitions", m.getPartitionsHandler)
	http.HandleFunc("/getPartitionById", m.getPartitionByIDHandler)
	http.HandleFunc("/getInode", m.getInodeHandler)
	// get all inodes of the partitionID
	http.HandleFunc("/getAllInodes", m.getAllInodesHandler)
	// get dentry information
	http.HandleFunc("/getDentry", m.getDentryHandler)
	http.HandleFunc("/getDirectory", m.getDirectoryHandler)
	http.HandleFunc("/getAllDentry", m.getAllDentriesHandler)
	http.HandleFunc("/status", m.getStatus)
	return
}
