package api

import (
	"fmt"
	"net/http"

	"github.com/go-openapi/runtime"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/api/v1/models"
)

type APIError struct {
	code int
	msg  string
}

// New creates a API error from the code, msg and extra arguments.
func New(code int, msg string, args ...interface{}) *APIError {
	if code <= 0 {
		code = 500
	}

	if len(args) > 0 {
		return &APIError{code: code, msg: fmt.Sprintf(msg, args...)}
	}

	return &APIError{code: code, msg: msg}
}

// Error creates a new API error from the code and error.
func Error(code int, err error) *APIError {
	if err == nil {
		err = fmt.Errorf("Error pointer was nil")
	}

	return New(code, err.Error())
}

// WriteResponse to the client.
func (a *APIError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
	rw.WriteHeader(a.code)
	m := a.GetModel()
	if err := producer.Produce(rw, m); err != nil {
		panic(err)
	}
}

func (a *APIError) GetModel() *models.Error {
	m := models.Error(a.msg)
	return &m
}
