package main

import (
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
	"gopkg.in/sohlich/elogrus.v7"
	"os"
	"testing"
)

func TestLog(test *testing.T) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.WarnLevel)
	log.SetOutput(os.Stdout)
	log.WithFields(log.Fields{
		"animal": "walrus",
	}).Info("A walrus appears")
	log.WithFields(log.Fields{
		"animal": "walrus",
	}).Warn("Warn: A walrus appears")
}

func TestELK(test *testing.T) {
	logger := log.New()
	client, err := elastic.NewClient(elastic.SetURL("http://10.202.4.125:20114"))
	if err != nil {
		logger.Panic(err)
	}
	hook, err := elogrus.NewAsyncElasticHook(client, "10.202.4.125", log.DebugLevel, "qssweb-business-log-fanyi_so_com")
	if err != nil {
		logger.Panic(err)
	}

	logger.Hooks.Add(hook)

	logger.WithFields(log.Fields{
		"name": "joe",
		"age":  42,
	}).Error("Hello world!")
}
