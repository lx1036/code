package server

import (
	"context"
	v1 "k8s-lx1036/k8s/kubelet/cri-hook-server/pkg/apis/crihookserver.k9s.io/v1"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gorilla/mux"
	"k8s.io/klog/v2"
)

type HookHandler interface {
	PreHook(ctx context.Context, patch *PatchData, method, path string, body []byte) error
	PostHook(ctx context.Context, patch *PatchData, method, path string, body []byte) error
}

type HookServer struct {
	mux           *mux.Router
	listenAddress string

	timeout time.Duration

	// docker server
	backend http.Handler
}

// hookHandleKey group http request options
type hookHandleKey struct {
	Method     string
	URLPattern string
}

// hookHandleData group kinds of hook handler
type hookHandleData struct {
	preHooks  []HookHandler
	postHooks []HookHandler
}

func NewHookServer(config *v1.HookConfiguration) *HookServer {
	server := &HookServer{
		mux:           mux.NewRouter(),
		timeout:       config.Timeout * time.Second,
		listenAddress: config.ListenAddress,
		backend:       newReverseProxy(config.RemoteEndpoint), // RemoteEndpoint: "unix:///var/run/docker.sock"
	}

	hooksMap := make(map[hookHandleKey]*hookHandleData)
	for _, webhook := range config.WebHooks {
		klog.Infof("Register hook %s, endpoint %s", webhook.Name, webhook.Endpoint)
		handler := newWebhookConnector(webhook.Name, webhook.Endpoint, webhook.FailurePolicy)
		for _, stage := range webhook.Stages {
			key := hookHandleKey{
				Method:     stage.Method,
				URLPattern: stage.URLPattern,
			}

			hookData, found := hooksMap[key]
			if !found {
				// must initialize the hook handler data
				hookData = &hookHandleData{
					preHooks:  make([]HookHandler, 0),
					postHooks: make([]HookHandler, 0),
				}
			}

			switch stage.Type {
			case v1.PreHookType:
				hookData.preHooks = append(hookData.preHooks, handler)
				hooksMap[key] = hookData
			case v1.PostHookType:
				hookData.postHooks = append(hookData.postHooks, handler)
				hooksMap[key] = hookData
			}
		}
	}

	for key, data := range hooksMap {
		preHookChainHandler := server.buildPreHookHandlerFunc(data.preHooks)
		postHookChainHandler := server.buildPostHookHandlerFunc(data.postHooks)

		server.mux.Methods(key.Method).Path(key.URLPattern).HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if err := preHookChainHandler(writer, request); err != nil {
				return
			}
			server.backend.ServeHTTP(writer, request)
			postHookChainHandler(writer, request)

			/*recorder := httptest.NewRecorder()
			server.backend.ServeHTTP(recorder, request)

			postHookChainHandler(recorder, request)

			for k, vs := range recorder.Header() {
				for _, v := range vs {
					writer.Header().Set(k, v)
				}
			}
			writer.WriteHeader(recorder.Code)
			writer.Write(recorder.Body.Bytes())*/
		})
	}

}

type PreHookFunc func(w http.ResponseWriter, r *http.Request) error
type PostHookFunc func(w *httptest.ResponseRecorder, r *http.Request)

func (server *HookServer) buildPreHookHandlerFunc(handlers []HookHandler) PreHookFunc {

	return func(w http.ResponseWriter, r *http.Request) error {

		if err := server.applyHook(ctx, handlers, v1.PreHookType, r.Method, r.URL.Path, &bodyBytes); err != nil {

		}

		return nil
	}
}

func (server *HookServer) buildPostHookHandlerFunc(handlers []HookHandler) PostHookFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if err := server.applyHook(ctx, handlers, v1.PreHookType, r.Method, r.URL.Path, &bodyBytes); err != nil {

		}

	}
}
