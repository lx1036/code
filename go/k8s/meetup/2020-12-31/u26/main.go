package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s-lx1036/k8s/meetup/2020-12-31/u26/collector"
	"k8s-lx1036/k8s/meetup/2020-12-31/u26/reporter"
	"k8s-lx1036/k8s/meetup/2020-12-31/u26/request"
	"k8s-lx1036/k8s/meetup/2020-12-31/u26/storage"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})

	stopCh := Signal()
	redisStorage := storage.NewRedisMetricsStorage()

	// 终端输出数据
	console := reporter.NewConsole(redisStorage)
	go console.StartRepeatedReport(stopCh, time.Now().Add(-10*time.Minute), time.Now())
	//邮件输出数据
	//email := reporter.NewMail(redisStorage)
	//email.AddToAddress([]string{"example@example.com"})
	//email.StartRepeatedReport(stopCh, time.Now().Add(-10 * time.Minute),time.Now())

	c := collector.NewMetricsCollector(redisStorage)
	c.RecordRequest(request.RequestInfo{
		ApiName:      "register",
		ResponseTime: time.Second * 1,
		Timestamp:    time.Now(),
	})
	c.RecordRequest(request.RequestInfo{
		ApiName:      "login",
		ResponseTime: time.Second * 2,
		Timestamp:    time.Now().Add(time.Second * 30),
	})

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
