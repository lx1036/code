package master

import (
	"encoding/json"
	"errors"
	"k8s.io/klog/v2"
	"net/http"
	"regexp"
	"strconv"

	"k8s-lx1036/k8s/storage/sunfs/pkg/proto"
)

var (
	volNameRegexp *regexp.Regexp
)

func (server *Server) getVolSimpleInfo(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		name    string
		vol     *Volume
		volView *proto.SimpleVolView
	)
	if name, err = parseAndExtractName(r); err != nil {
		sendErrReply(w, r, &proto.HTTPReply{Code: proto.ErrCodeParamError, Msg: err.Error()})
		return
	}
	if vol, err = server.cluster.getVolume(name); err != nil {
		sendErrReply(w, r, newErrHTTPReply(proto.ErrVolNotExists))
		return
	}
	volView = &proto.SimpleVolView{
		ID:            vol.ID,
		Name:          vol.Name,
		Owner:         vol.Owner,
		MpReplicaNum:  vol.mpReplicaNum,
		Status:        vol.Status,
		Capacity:      vol.Capacity,
		MpCnt:         len(vol.MetaPartitions),
		S3Endpoint:    vol.s3Endpoint,
		BucketDeleted: vol.bucketdeleted,
	}

	sendOkReply(w, r, newSuccessHTTPReply(volView))
}

func parseAndExtractName(r *http.Request) (name string, err error) {
	if err = r.ParseForm(); err != nil {
		return
	}
	return extractName(r)
}

func extractName(r *http.Request) (name string, err error) {
	if name = r.FormValue(nameKey); name == "" {
		err = keyNotFound(nameKey)
		return
	}
	if !volNameRegexp.MatchString(name) {
		return "", errors.New("name can only be number and letters")
	}

	return
}
func sendOkReply(w http.ResponseWriter, r *http.Request, httpReply *proto.HTTPReply) (err error) {
	switch httpReply.Data.(type) {
	case *MetaPartition:
		mp := httpReply.Data.(*MetaPartition)
		mp.RLock()
		defer mp.RUnlock()
	case *MetaNode:
		mn := httpReply.Data.(*MetaNode)
		mn.RLock()
		defer mn.RUnlock()
	}
	reply, err := json.Marshal(httpReply)
	if err != nil {
		klog.Errorf("fail to marshal http reply[%v]. URL[%v],remoteAddr[%v] err:[%v]", httpReply, r.URL, r.RemoteAddr, err)
		return
	}

	send(w, r, reply)
	return
}

func sendErrReply(w http.ResponseWriter, r *http.Request, httpReply *proto.HTTPReply) {
	klog.Infof("URL[%v],remoteAddr[%v],response err[%v]", r.URL, r.RemoteAddr, httpReply)
	reply, err := json.Marshal(httpReply)
	if err != nil {
		klog.Errorf("fail to marshal http reply[%v]. URL[%v],remoteAddr[%v] err:[%v]", httpReply, r.URL, r.RemoteAddr, err)
		http.Error(w, "fail to marshal http reply", http.StatusBadRequest)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(reply)))
	if _, err = w.Write(reply); err != nil {
		klog.Errorf("fail to write http reply[%s] len[%d].URL[%v],remoteAddr[%v] err:[%v]", string(reply), len(reply), r.URL, r.RemoteAddr, err)
	}
	return
}

func newErrHTTPReply(err error) *proto.HTTPReply {
	if err == nil {
		return newSuccessHTTPReply("")
	}
	code, ok := proto.Err2CodeMap[err]
	if ok {
		return &proto.HTTPReply{Code: code, Msg: err.Error()}
	}
	return &proto.HTTPReply{Code: proto.ErrCodeInternalError, Msg: err.Error()}
}

func newSuccessHTTPReply(data interface{}) *proto.HTTPReply {
	return &proto.HTTPReply{Code: proto.ErrCodeSuccess, Msg: proto.ErrSuc.Error(), Data: data}
}

func send(w http.ResponseWriter, r *http.Request, reply []byte) {
	w.Header().Set("content-type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(reply)))
	if _, err := w.Write(reply); err != nil {
		klog.Errorf("fail to write http reply[%s] len[%d].URL[%v],remoteAddr[%v] err:[%v]", string(reply), len(reply), r.URL, r.RemoteAddr, err)
		return
	}
	klog.Infof("URL[%v],remoteAddr[%v],response ok", r.URL, r.RemoteAddr)
	return
}
