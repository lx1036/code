package main

/**
https://github.com/xianlubird/mydocker/blob/master/main.go
*/
import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/urfave/cli"
	"os"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:    "my-docker",
		Short:  "write my own docker",
		PreRun: preRun,
		Run:    run,
	}

	rootCmd.AddCommand(cmd.InitCmd)

	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}

func preRun(cmd *cobra.Command, args []string) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
}

func run(cmd *cobra.Command, args []string) {

}

func mainlx() {
	app := cli.NewApp()
	app.Name = "my-docker"
	app.Usage = `write my docker`

	app.Commands = []cli.Command{
		//initCommand,
		runCommand,
		listCommand,
		logCommand,
		execCommand,
		stopCommand,
		removeCommand,
		commitCommand,
		networkCommand,
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

var runCommand = cli.Command{
	Name:  "run",
	Usage: `Create a container with namespace and cgroups limit ie: mydocker run -ti [image] [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		},
	},
	Action: func(context *cli.Context) error {

		//cmd := context.Args().Get(0)
		//tty := context.Bool("ti")

		//Run(tty, cmd)
		return nil
	},
}

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list all the containers",
	Action: func(context *cli.Context) error {
		ListContainers()
		return nil
	},
}

func ListContainers() {

}

var logCommand = cli.Command{
	Name:  "logs",
	Usage: "print logs of a container",
	Action: func(context *cli.Context) error {

		return nil
	},
}

var execCommand = cli.Command{
	Name:  "exec",
	Usage: "exec a command into container",
	Action: func(context *cli.Context) error {
		return nil
	},
}

var stopCommand = cli.Command{
	Name:  "stop",
	Usage: "stop a container",
	Action: func(context *cli.Context) error {
		return nil
	},
}

var removeCommand = cli.Command{
	Name:  "rm",
	Usage: "remove unused containers",
	Action: func(context *cli.Context) error {
		return nil
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into image",
	Action: func(context *cli.Context) error {
		return nil
	},
}

var networkCommand = cli.Command{
	Name:  "network",
	Usage: "container network commands",
	Subcommands: []cli.Command{
		{
			Name:  "create",
			Usage: "create a container network",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "driver",
					Usage: "network driver",
				},
				cli.StringFlag{
					Name:  "subnet",
					Usage: "subnet cidr",
				},
			},
			Action: func(context *cli.Context) error {
				return nil
			},
		},
	},
}
