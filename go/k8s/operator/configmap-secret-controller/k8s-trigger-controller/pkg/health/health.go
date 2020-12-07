package health

import (
	"github.com/sirupsen/logrus"
	"net/http"
)

type HealthCheckHandler struct{}

func (h *HealthCheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("ok"))
}

func SetupHealthCheckHandler() {
	go func() {
		http.Handle("/healthz", &HealthCheckHandler{})
		logrus.Fatal(http.ListenAndServe(":9001", nil))
	}()
}
