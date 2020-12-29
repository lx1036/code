package main

import (
	"os"
	"os/signal"
	"syscall"
	
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})
	
	stopCh := Signal()
	
	redisStorage := storage.NewRedis()
	
	console := reporter.NewConsole(redisStorage)
	console.StartRepeatedReport(60,60)
	email := reporter.NewMail(redisStorage)
	email.AddToAddress([]string{"example@example.com"})
	email.StartDailyReport()
	
	collector := metrics.NewCollector(redisStorage)
	collector.RecordRequest(metrics.RequestInfo{})
	
	
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
