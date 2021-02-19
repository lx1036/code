package pkg

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"go.etcd.io/etcd/raft/raftpb"
	"k8s.io/klog/v2"
)

// Handler for a http based key-value store backed by raft
type httpKVAPI struct {
	store       *KVStore
	confChangeC chan<- raftpb.ConfChange
}

func (h *httpKVAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.RequestURI
	defer r.Body.Close()

	switch {
	case r.Method == http.MethodPut:
		v, err := ioutil.ReadAll(r.Body)
		if err != nil {
			klog.Errorf("Failed to read on PUT (%v)\n", err)
			http.Error(w, "Failed on PUT", http.StatusBadRequest)
			return
		}

		h.store.Propose(key, string(v))

		// Optimistic-- no waiting for ack from raft. Value is not yet
		// committed so a subsequent GET on the key may return old value
		w.WriteHeader(http.StatusNoContent)
	case r.Method == "GET":
		if v, ok := h.store.Lookup(key); ok {
			w.Write([]byte(v))
		} else {
			http.Error(w, "Failed to GET", http.StatusNotFound)
		}
	default:
		w.Header().Set("Allow", "PUT")
		w.Header().Add("Allow", "GET")
		w.Header().Add("Allow", "POST")
		w.Header().Add("Allow", "DELETE")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

// serveHttpKVAPI starts a key-value server with a GET/PUT API and listens.
func ServeHttpKVAPI(kv *KVStore, port int, confChangeC chan<- raftpb.ConfChange, errorC <-chan error) {
	srv := http.Server{
		Addr: ":" + strconv.Itoa(port),
		Handler: &httpKVAPI{
			store:       kv,
			confChangeC: confChangeC,
		},
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			klog.Fatal(err)
		}
	}()

	// exit when raft goes down
	if err, ok := <-errorC; ok {
		klog.Fatal(err)
	}
}
