package master

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"

	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
	"k8s-lx1036/k8s/storage/fusefs/pkg/util"

	"github.com/gorilla/mux"
	"k8s.io/klog/v2"
)

const (
	ParamsOwner    = "owner"
	ParamsName     = "name"
	ParamsCapacity = "capacity"
	ParamsAddrKey  = "addr"
)

func (server *Server) startHTTPService() {
	router := mux.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if server.isRaftLeader() {
				next.ServeHTTP(writer, request)
				return
			}

			if len(server.leaderInfo.addr) == 0 {
				http.Error(writer, fmt.Sprintf("no raft leader"), http.StatusBadRequest)
				return
			}

			// proxy request to raft leader
			reverseProxy := &httputil.ReverseProxy{
				Director: func(request *http.Request) {
					request.URL.Scheme = "http"
					request.URL.Host = server.leaderInfo.addr
				},
			}
			reverseProxy.ServeHTTP(writer, request)
		})
	})

	// cluster
	router.NewRoute().Methods(http.MethodGet).Path("/cluster/info").HandlerFunc(server.getClusterInfo)

	// meta node
	router.NewRoute().Methods(http.MethodPost).Path("/metanode").HandlerFunc(server.addMetaNode)
	router.NewRoute().Methods(http.MethodGet).Path("/metanode").HandlerFunc(server.getMetaNode)

	// volume
	router.NewRoute().Methods(http.MethodGet).Path("/vols").HandlerFunc(server.listVols)
	router.NewRoute().Methods(http.MethodGet).Path("/vol").HandlerFunc(server.getVol)
	router.NewRoute().Methods(http.MethodGet).Path("/vol/stat").HandlerFunc(server.getVolStat)
	router.NewRoute().Methods(http.MethodPost).Path("/vol").HandlerFunc(server.createVol)
	router.NewRoute().Methods(http.MethodPost).Path("/vol/expand").HandlerFunc(server.updateVol)
	router.NewRoute().Methods(http.MethodPost).Path("/vol/shrink").HandlerFunc(server.updateVol)

	// meta partition
	router.NewRoute().Methods(http.MethodPost).Path("/metapartition/expand").HandlerFunc(server.createMetaPartition)

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf("%s:%d", server.ip, server.port), router); err != nil {
			klog.Fatal(err)
		}
	}()
}

// cluster
func (server *Server) getClusterInfo(writer http.ResponseWriter, request *http.Request) {
	data, _ := json.Marshal(&proto.ClusterInfo{Cluster: server.cluster.Name, Ip: strings.Split(request.RemoteAddr, ":")[0]})
	send(writer, http.StatusOK, data)
	return
}

// volume
func (server *Server) listVols(writer http.ResponseWriter, request *http.Request) {
	vols := server.cluster.allVols()
	data, _ := json.Marshal(vols)
	send(writer, http.StatusOK, data)
	return
}

func (server *Server) getVol(writer http.ResponseWriter, request *http.Request) {
	name := request.FormValue(ParamsName)
	vol, err := server.cluster.getVolume(name)
	if err != nil {
		http.Error(writer, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}

	data, _ := json.Marshal(vol)
	send(writer, http.StatusOK, data)
	return
}

type VolStatInfo struct {
	Name        string
	TotalSize   uint64
	UsedSize    uint64
	UsedRatio   string
	EnableToken bool
	InodeCount  uint64
	Status      VolStatus
}

func (server *Server) getVolStat(writer http.ResponseWriter, request *http.Request) {
	name := request.FormValue(ParamsName)
	vol, err := server.cluster.getVolume(name)
	if err != nil {
		http.Error(writer, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}

	volStatInfo := &VolStatInfo{
		Name:       vol.Name,
		TotalSize:  vol.Capacity * util.GB,
		UsedSize:   vol.totalUsedSpace(),
		UsedRatio:  vol.UsedRatio,
		InodeCount: 0,
		Status:     vol.Status,
	}
	for _, metaPartition := range vol.MetaPartitions {
		volStatInfo.InodeCount += metaPartition.InodeCount
	}
	data, _ := json.Marshal(volStatInfo)
	send(writer, http.StatusOK, data)
	return
}

func (server *Server) createVol(writer http.ResponseWriter, request *http.Request) {
	name := request.FormValue(ParamsName)
	owner := request.FormValue(ParamsOwner)
	capacityStr := request.FormValue(ParamsCapacity)
	capacity, err := strconv.ParseUint(capacityStr, 10, 64)
	if err != nil {
		http.Error(writer, fmt.Sprintf("capacity params %s is wrong", capacityStr), http.StatusBadRequest)
		return
	}

	if vol, err := server.cluster.createVol(name, owner, capacity); err != nil {
		http.Error(writer, fmt.Sprintf("create volume %s err: %v", name, err), http.StatusInternalServerError)
		return
	} else {
		data, _ := json.Marshal(vol)
		send(writer, http.StatusOK, data)
		return
	}
}

// expand volume for fusefs csi
func (server *Server) updateVol(writer http.ResponseWriter, request *http.Request) {
	name := request.FormValue(ParamsName)
	owner := request.FormValue(ParamsOwner)
	capacityStr := request.FormValue(ParamsCapacity)
	capacity, err := strconv.ParseUint(capacityStr, 10, 64)
	if err != nil {
		http.Error(writer, fmt.Sprintf("capacity params %s is wrong", capacityStr), http.StatusBadRequest)
		return
	}

	if vol, err := server.cluster.updateVol(name, owner, capacity); err != nil {
		http.Error(writer, fmt.Sprintf("create volume %s err: %v", name, err), http.StatusInternalServerError)
		return
	} else {
		data, _ := json.Marshal(vol)
		send(writer, http.StatusOK, data)
		return
	}
}

// metanode
func (server *Server) addMetaNode(writer http.ResponseWriter, request *http.Request) {
	nodeAddr := request.FormValue(ParamsAddrKey)
	if metaNodeID, err := server.cluster.addMetaNode(nodeAddr); err != nil {
		http.Error(writer, fmt.Sprintf("add metanode %s err: %v", nodeAddr, err), http.StatusInternalServerError)
		return
	} else {
		send(writer, http.StatusOK, []byte(strconv.FormatUint(metaNodeID, 10)))
		return
	}
}

func (server *Server) getMetaNode(writer http.ResponseWriter, request *http.Request) {
	nodeAddr := request.FormValue(ParamsAddrKey)
	if metaNode, err := server.cluster.getMetaNode(nodeAddr); err != nil {
		http.Error(writer, fmt.Sprintf("add metanode %s err: %v", nodeAddr, err), http.StatusInternalServerError)
		return
	} else {
		metaNode.PersistenceMetaPartitions = server.cluster.getAllMetaPartitionIDByMetaNode(nodeAddr)
		data, _ := json.Marshal(metaNode)
		send(writer, http.StatusOK, data)
		return
	}
}

// meta partition
func (server *Server) createMetaPartition(writer http.ResponseWriter, request *http.Request) {

}

func send(writer http.ResponseWriter, code int, data []byte) {
	writer.Header().Set("content-type", "application/json")
	writer.Header().Set("Content-Length", strconv.Itoa(len(data)))
	writer.WriteHeader(code)
	_, _ = writer.Write(data)
	return
}
