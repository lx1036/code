package kube_gin

import (
	"bufio"
	"net"
	"net/http"
)

type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier

	// Returns the HTTP response status code of the current request.
	Status() int

	// Returns the number of bytes already written into the response http body.
	// See Written()
	Size() int

	// Writes the string into the response body.
	WriteString(string) (int, error)

	// Returns true if the response body was already written.
	Written() bool

	// Forces to write the http header (status code + headers).
	WriteHeaderNow()

	// get the http.Pusher for server push
	Pusher() http.Pusher
}

type responseWriter struct {
	http.ResponseWriter
	size   int
	status int
}

func (w responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	panic("implement me")
}

func (w responseWriter) Flush() {
	panic("implement me")
}

func (w responseWriter) CloseNotify() <-chan bool {
	panic("implement me")
}

func (w responseWriter) Status() int {
	return w.status
}

func (w responseWriter) Size() int {
	panic("implement me")
}

func (w responseWriter) WriteString(string) (int, error) {
	panic("implement me")
}

func (w responseWriter) Written() bool {
	panic("implement me")
}

func (w responseWriter) WriteHeaderNow() {
	panic("implement me")
}

func (w responseWriter) Pusher() http.Pusher {
	panic("implement me")
}

var _ ResponseWriter = &responseWriter{}
