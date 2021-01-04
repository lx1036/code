package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

var (
	stats = map[string]map[string]interface{}{}
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})

	stopCh := Signal()

	metrics := NewMetrics(stopCh)
	controller := UserController{metrics: metrics}
	controller.register()
	controller.register()
	controller.login("admin1", "password1")
	controller.login("admin2", "password2")

	<-stopCh
}

func Signal() (stopCh <-chan struct{}) {
	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1)
	}()

	return stop
}
