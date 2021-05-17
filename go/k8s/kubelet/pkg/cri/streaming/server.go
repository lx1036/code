package streaming

import (
	"net/http"

	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

// Server is the library interface to serve the stream requests.
type Server interface {
	http.Handler

	// Get the serving URL for the requests.
	// Requests must not be nil. Responses may be nil iff an error is returned.
	GetExec(*runtimeapi.ExecRequest) (*runtimeapi.ExecResponse, error)
	GetAttach(req *runtimeapi.AttachRequest) (*runtimeapi.AttachResponse, error)
	GetPortForward(*runtimeapi.PortForwardRequest) (*runtimeapi.PortForwardResponse, error)

	// Start the server.
	// addr is the address to serve on (address:port) stayUp indicates whether the server should
	// listen until Stop() is called, or automatically stop after all expected connections are
	// closed. Calling Get{Exec,Attach,PortForward} increments the expected connection count.
	// Function does not return until the server is stopped.
	Start(stayUp bool) error
	// Stop the server, and terminate any open connections.
	Stop() error
}
