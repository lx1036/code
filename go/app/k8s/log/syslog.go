package main

import (
	"log"
	"log/syslog"
)

func main() {
	// tail -f /var/log/system.log
	logwriter, e := syslog.New(syslog.LOG_NOTICE, "mygolangprogram")
	if e == nil {
		log.SetOutput(logwriter)
	}

	log.Print("Hello worldogs!")
}
