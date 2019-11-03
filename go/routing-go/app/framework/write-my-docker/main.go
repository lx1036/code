package main

/**
https://github.com/xianlubird/mydocker/blob/master/main.go
 */
import (
	"github.com/urfave/cli"
	log "github.com/sirupsen/logrus"
	"os"
)

const usage  = `mydocker is a simple container runtime implementation. 
	The purpose of this project is to learn how docker works and how to write a docker by ourselves
	Enjoy it, just for fun.`

func main()  {
	app := cli.NewApp()
	app.Name = "my-docker"
	app.Usage = usage

	app.Commands = []cli.Command{
		initCommand,
		runCommand,
	}

	app.Before = func(context *cli.Context) error {
		log.SetFormatter(&log.JSONFormatter{}) // &log.JSONFormatter === *JSONFormatter, Format() belongs to *JSONFormatter
		log.SetOutput(os.Stdout)

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
