package server

import (
    "context"
    "github.com/cilium/cilium/api/v1/server/restapi"
    "os"
)

var (
    // ServerCtx and ServerCancel
    ServerCtx, serverCancel = context.WithCancel(context.Background())
)

const (
    schemeHTTP  = "http"
    schemeHTTPS = "https"
    schemeUnix  = "unix"
)

var defaultSchemes []string

func init() {
    defaultSchemes = []string{
        schemeUnix,
    }
}

type Server struct {
}

func NewServer(api *restapi.CiliumAPIAPI) *Server {
    s := new(Server)

    s.shutdown = make(chan struct{})
    s.api = api
    s.interrupt = make(chan os.Signal, 1)
    return s
}
