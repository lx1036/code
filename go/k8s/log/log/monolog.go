package main

import (
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
	syslogHandler "github.com/sirupsen/logrus/hooks/syslog"
	"github.com/spf13/viper"
	"gopkg.in/sohlich/elogrus.v7"
	"log/syslog"
	"os"
)

func main() {
	viper.AutomaticEnv()

	log.SetFormatter(&log.JSONFormatter{})
	if viper.GetBool("debug") {
		log.SetLevel(log.TraceLevel)
	} else {
		log.SetLevel(log.InfoLevel) // >= only write info level
	}

	log.SetOutput(os.Stdout) // add stdout handler

	syslogHook, err := syslogHandler.NewSyslogHook("", "", syslog.LOG_INFO, "syslog")
	if err != nil {
		log.Panic(err)
	}
	log.AddHook(syslogHook) // add syslog handler (tail -f /var/log/system.log in mac)

	client, err := elastic.NewClient(elastic.SetURL(viper.GetString("ELASTIC_URL")))
	if err != nil {
		log.Panic(err)
	}
	EfkHook, err := elogrus.NewAsyncElasticHook(client, viper.GetString("EFK_URL"), log.DebugLevel, "lx1036-account")
	if err != nil {
		log.Panic(err)
	}
	log.AddHook(EfkHook) // add efk handler

	log.WithFields(log.Fields{
		"app_level":  "bug",
		"message":    "user-agent from iphone",
		"account_id": 1,
	}).Info("efk and stdout") // write log
}
