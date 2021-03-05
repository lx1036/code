package master

import (
	"k8s-lx1036/k8s/storage/dfs/pkg/proto"
	"net/http"

	"k8s.io/klog/v2"
)

func (m *Server) startHTTPService() {
	go func() {
		m.handleFunctions()
		if err := http.ListenAndServe(colonSplit+m.port, nil); err != nil {
			klog.Errorf("action[startHTTPService] failed,err[%v]", err)
			panic(err)
		}
	}()
	return
}

func (m *Server) handleFunctions() {
	http.HandleFunc(proto.AdminGetIP, m.getIPAddr)
	http.Handle(proto.AdminGetCluster, m.handlerWithInterceptor())
	http.Handle(proto.AdminGetVolMountClient, m.handlerWithInterceptor())
	http.Handle(proto.AdminCreateVol, m.handlerWithInterceptor())
	http.Handle(proto.AdminGetVol, m.handlerWithInterceptor())
	http.Handle(proto.AdminDeleteVol, m.handlerWithInterceptor())
	http.Handle(proto.AdminUpdateVol, m.handlerWithInterceptor())
	http.Handle(proto.AdminClusterFreeze, m.handlerWithInterceptor())
	http.Handle(proto.AddMetaNode, m.handlerWithInterceptor())
	http.Handle(proto.DecommissionMetaNode, m.handlerWithInterceptor())
	http.Handle(proto.GetMetaNode, m.handlerWithInterceptor())
	http.Handle(proto.AdminLoadMetaPartition, m.handlerWithInterceptor())
	http.Handle(proto.AdminDecommissionMetaPartition, m.handlerWithInterceptor())
	http.Handle(proto.AdminAddMetaReplica, m.handlerWithInterceptor())
	http.Handle(proto.AdminDeleteMetaReplica, m.handlerWithInterceptor())
	http.Handle(proto.ClientVol, m.handlerWithInterceptor())
	http.Handle(proto.ClientMetaPartitions, m.handlerWithInterceptor())
	http.Handle(proto.ClientMetaPartition, m.handlerWithInterceptor())
	http.Handle(proto.GetMetaNodeTaskResponse, m.handlerWithInterceptor())
	http.Handle(proto.AdminCreateMP, m.handlerWithInterceptor())
	http.Handle(proto.ClientVolStat, m.handlerWithInterceptor())
	http.Handle(proto.ClientVolMount, m.handlerWithInterceptor())
	http.Handle(proto.ClientVolMountUpdate, m.handlerWithInterceptor())
	http.Handle(proto.ClientVolUnMount, m.handlerWithInterceptor())
	http.Handle(proto.AddRaftNode, m.handlerWithInterceptor())
	http.Handle(proto.RemoveRaftNode, m.handlerWithInterceptor())
	http.Handle(proto.AdminSetMetaNodeThreshold, m.handlerWithInterceptor())
	http.Handle(proto.GetTopologyView, m.handlerWithInterceptor())

	return
}
