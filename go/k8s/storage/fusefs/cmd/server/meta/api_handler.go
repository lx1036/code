package meta

import (
	"encoding/json"
	"net/http"
	"sync/atomic"

	"k8s.io/klog/v2"
)

// APIResponse defines the structure of the response to an HTTP request
type APIResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// Marshal is a wrapper function of json.Marshal
func (api *APIResponse) Marshal() ([]byte, error) {
	return json.Marshal(api)
}

// NewAPIResponse returns a new API response.
func NewAPIResponse(code int, msg string) *APIResponse {
	return &APIResponse{
		Code: code,
		Msg:  msg,
	}
}

// register the APIs
func (m *MetaNode) registerAPIHandler() (err error) {
	http.HandleFunc("/getPartitions", m.getPartitionsHandler)
	//http.HandleFunc("/getPartitionById", m.getPartitionByIDHandler)
	//http.HandleFunc("/getInode", m.getInodeHandler)
	// get all inodes of the partitionID
	//http.HandleFunc("/getAllInodes", m.getAllInodesHandler)
	// get dentry information
	//http.HandleFunc("/getDentry", m.getDentryHandler)
	//http.HandleFunc("/getDirectory", m.getDirectoryHandler)
	//http.HandleFunc("/getAllDentry", m.getAllDentriesHandler)
	//http.HandleFunc("/status", m.getStatus)
	return
}

func (m *MetaNode) getPartitionsHandler(w http.ResponseWriter, r *http.Request) {
	resp := NewAPIResponse(http.StatusOK, http.StatusText(http.StatusOK))
	defer func() {
		data, _ := resp.Marshal()
		if _, err := w.Write(data); err != nil {
			klog.Errorf("[getPartitionsHandler] response %v", err)
		}
	}()

	if atomic.LoadUint32(&m.state) != StateRunning {
		resp.Code = http.StatusBadRequest
		resp.Msg = "server not running"
		return
	}
	resp.Data = m.metadataManager
}
