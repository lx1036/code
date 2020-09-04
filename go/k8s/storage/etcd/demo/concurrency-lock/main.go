package main

import (
	"context"
	"go.etcd.io/etcd/clientv3"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3/concurrency"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// https://segmentfault.com/a/1190000021603215
func main() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 10,
	})
	if err != nil {
		panic(err)
	}

	lockKey := "/lock"

	go func() {
		session, err := concurrency.NewSession(client)
		if err != nil {
			log.WithFields(log.Fields{
				"errmsg": err.Error(),
			}).Error("[job2 session]")
		}

		mutex := concurrency.NewMutex(session, lockKey)
		err = mutex.Lock(context.TODO())
		if err != nil {
			log.WithFields(log.Fields{
				"errmsg": err.Error(),
			}).Error("[job1 mutex]")
		}
		defer mutex.Unlock(context.TODO())

		// business logic
		log.WithFields(log.Fields{
			"msg": "do job1",
		}).Info("[job1]")
		time.Sleep(time.Second * 3)
	}()

	go func() {
		session, err := concurrency.NewSession(client)
		if err != nil {
			log.WithFields(log.Fields{
				"errmsg": err.Error(),
			}).Error("[job2 session]")
		}

		mutex := concurrency.NewMutex(session, lockKey)
		err = mutex.Lock(context.TODO())
		if err != nil {
			log.WithFields(log.Fields{
				"errmsg": err.Error(),
			}).Error("[job2 mutex]")
		}
		defer mutex.Unlock(context.TODO())

		// business logic
		log.WithFields(log.Fields{
			"msg": "do job2",
		}).Info("[job2]")
		time.Sleep(time.Second * 3)
	}()

	<-sig
}
