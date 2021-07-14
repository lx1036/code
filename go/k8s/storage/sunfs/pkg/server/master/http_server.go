package master

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"k8s-lx1036/k8s/storage/sunfs/pkg/util/proto"

	"k8s.io/klog/v2"
)

func (server *Server) startHTTPService() {
	go func() {
		server.handleFunctions()
		if err := http.ListenAndServe(colonSplit+server.port, nil); err != nil {
			klog.Errorf("action[startHTTPService] failed,err[%v]", err)
			panic(err)
		}
	}()
	return
}

func (server *Server) handleFunctions() {
	//http.HandleFunc(proto.AdminGetIP, server.getIPAddr)
	http.Handle(proto.AdminGetCluster, server.handlerWithInterceptor())
	http.Handle(proto.AdminGetVolMountClient, server.handlerWithInterceptor())
	http.Handle(proto.AdminCreateVol, server.handlerWithInterceptor())
	http.Handle(proto.AdminGetVol, server.handlerWithInterceptor())
	http.Handle(proto.AdminDeleteVol, server.handlerWithInterceptor())
	http.Handle(proto.AdminUpdateVol, server.handlerWithInterceptor())
	http.Handle(proto.AdminClusterFreeze, server.handlerWithInterceptor())
	http.Handle(proto.AddMetaNode, server.handlerWithInterceptor())
	http.Handle(proto.DecommissionMetaNode, server.handlerWithInterceptor())
	http.Handle(proto.GetMetaNode, server.handlerWithInterceptor())
	http.Handle(proto.AdminLoadMetaPartition, server.handlerWithInterceptor())
	http.Handle(proto.AdminDecommissionMetaPartition, server.handlerWithInterceptor())
	http.Handle(proto.AdminAddMetaReplica, server.handlerWithInterceptor())
	http.Handle(proto.AdminDeleteMetaReplica, server.handlerWithInterceptor())
	http.Handle(proto.ClientVol, server.handlerWithInterceptor())
	http.Handle(proto.ClientMetaPartitions, server.handlerWithInterceptor())
	http.Handle(proto.ClientMetaPartition, server.handlerWithInterceptor())
	http.Handle(proto.GetMetaNodeTaskResponse, server.handlerWithInterceptor())
	http.Handle(proto.AdminCreateMP, server.handlerWithInterceptor())
	http.Handle(proto.ClientVolStat, server.handlerWithInterceptor())
	http.Handle(proto.ClientVolMount, server.handlerWithInterceptor())
	http.Handle(proto.ClientVolMountUpdate, server.handlerWithInterceptor())
	http.Handle(proto.ClientVolUnMount, server.handlerWithInterceptor())
	http.Handle(proto.AddRaftNode, server.handlerWithInterceptor())
	http.Handle(proto.RemoveRaftNode, server.handlerWithInterceptor())
	http.Handle(proto.AdminSetMetaNodeThreshold, server.handlerWithInterceptor())
	http.Handle(proto.GetTopologyView, server.handlerWithInterceptor())
}

func (server *Server) handlerWithInterceptor() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if server.partition.IsRaftLeader() {
				if server.metaReady {
					server.ServeHTTP(w, r)
					return
				}
				klog.Warningf("action[handlerWithInterceptor] leader meta has not ready")
				http.Error(w, server.leaderInfo.addr, http.StatusBadRequest)
				return
			}
			if server.leaderInfo.addr == "" {
				klog.Errorf("action[handlerWithInterceptor] no leader,request[%v]", r.URL)
				http.Error(w, "no leader", http.StatusBadRequest)
				return
			}

			server.reverseProxy.ServeHTTP(w, r)

		})
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	klog.Infof("URL[%v],remoteAddr[%v]", r.URL, r.RemoteAddr)
	switch r.URL.Path {
	/*case proto.AdminGetCluster:
		server.getCluster(w, r)
	case proto.AdminGetVolMountClient:
		server.getVolMountClient(w, r)
	case proto.AdminCreateVol:
		server.createVol(w, r)*/
	case proto.AdminGetVol:
		server.getVolSimpleInfo(w, r)
	/*case proto.AdminDeleteVol:
		server.markDeleteVol(w, r)
	case proto.AdminUpdateVol:
		server.updateVol(w, r)
	case proto.AdminClusterFreeze:
		server.setupAutoAllocation(w, r)
	case proto.AddMetaNode:
		server.addMetaNode(w, r)
	case proto.GetMetaNode:
		server.getMetaNode(w, r)
	case proto.DecommissionMetaNode:
		server.decommissionMetaNode(w, r)
	case proto.GetMetaNodeTaskResponse:
		server.handleMetaNodeTaskResponse(w, r)
	case proto.ClientVol:
		server.getVol(w, r)
	case proto.ClientMetaPartitions:
		server.getMetaPartitions(w, r)
	case proto.ClientMetaPartition:
		server.getMetaPartition(w, r)
	case proto.ClientVolStat:
		server.getVolStatInfo(w, r)
	case proto.ClientVolMount:
		server.createVolMountClient(w, r)
	case proto.ClientVolMountUpdate:
		server.updateVolMountClientInfo(w, r)
	case proto.ClientVolUnMount:
		server.deleteVolMountClient(w, r)
	case proto.AdminLoadMetaPartition:
		server.loadMetaPartition(w, r)
	case proto.AdminDecommissionMetaPartition:
		server.decommissionMetaPartition(w, r)
	case proto.AdminCreateMP:
		server.createMetaPartition(w, r)
	case proto.AdminAddMetaReplica:
		server.addMetaReplica(w, r)
	case proto.AdminDeleteMetaReplica:
		server.deleteMetaReplica(w, r)
	case proto.AddRaftNode:
		server.addRaftNode(w, r)
	case proto.RemoveRaftNode:
		server.removeRaftNode(w, r)
	case proto.AdminSetMetaNodeThreshold:
		server.setMetaNodeThreshold(w, r)
	case proto.GetTopologyView:
		server.getTopology(w, r)*/
	default:
		http.Error(w, fmt.Sprintf("unsupported url %s", r.URL.String()), http.StatusBadRequest)
	}
}

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

func (server *Server) newReverseProxy() *httputil.ReverseProxy {
	return &httputil.ReverseProxy{Director: func(request *http.Request) {
		request.URL.Scheme = "http"
		request.URL.Host = server.leaderInfo.addr
	}}
}
