package kube_gin

import (
	"fmt"
	"net/http"
	"testing"
)

func TestEngine_Step1(test *testing.T) {
	engine := New()
	engine.Get("/", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprintf(writer, "ok\n")
	})

	_ = engine.Run(":9999")
}
