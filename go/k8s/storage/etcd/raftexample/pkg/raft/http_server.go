package raft

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"

	"github.com/gorilla/mux"

	"k8s.io/klog/v2"
)

func (server *Server) startHTTPService(port int) {
	router := mux.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if server.isLeader() {
				next.ServeHTTP(writer, request)
				return
			}

			leaderAddr := server.getLeaderAddr()
			if len(leaderAddr) == 0 {
				http.Error(writer, fmt.Sprintf("no raft leader"), http.StatusBadRequest)
				return
			}
			klog.Infof(fmt.Sprintf("new leader addr: %s", leaderAddr))

			// proxy request to raft leader
			reverseProxy := &httputil.ReverseProxy{
				Director: func(request *http.Request) {
					request.URL.Scheme = "http"
					request.URL.Host = leaderAddr
				},
			}
			reverseProxy.ServeHTTP(writer, request)
		})
	})

	router.NewRoute().Methods(http.MethodPost).Path("/vol").HandlerFunc(server.createVol)

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), router); err != nil {
			klog.Fatal(err)
		}
	}()
}

func (server *Server) createVol(writer http.ResponseWriter, request *http.Request) {
	name := request.FormValue("name")
	capacity := request.FormValue("capacity")
	if vol, err := server.cluster.createVol(name, capacity); err != nil {
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
