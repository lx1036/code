package master

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"

	"github.com/gorilla/mux"
	"k8s.io/klog/v2"
)

const (
	ParamsOwner    = "owner"
	ParamsName     = "name"
	ParamsCapacity = "capacity"
)

func (server *Server) startHTTPService() {
	router := mux.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if server.partition.IsRaftLeader() {
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
				Transport:      nil,
				FlushInterval:  0,
				ErrorLog:       nil,
				BufferPool:     nil,
				ModifyResponse: nil,
				ErrorHandler:   nil,
			}
			reverseProxy.ServeHTTP(writer, request)
		})
	})

	// volume
	router.NewRoute().Methods(http.MethodPost).Path("/vol").HandlerFunc(server.createVol)

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf("%s:%d", server.ip, server.port), router); err != nil {
			klog.Fatal(err)
		}
	}()
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

func send(writer http.ResponseWriter, code int, data []byte) {
	writer.Header().Set("content-type", "application/json")
	writer.Header().Set("Content-Length", strconv.Itoa(len(data)))
	writer.WriteHeader(code)
	_, _ = writer.Write(data)
	return
}
