package cmd

import (
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"
	"os"
)

var RootCmd = &cobra.Command{
	Use:                        "k8s_watcher",
	Short:                      "a watcher for k8s resource",
	Long:                       `
	k8s_watcher: a watcher for k8s resource.
	supported sink:
		- slack
		- email
		- webhook
`,
	Run: func(cmd *cobra.Command, args []string) {
		
	},
}

func Execute()  {
	if err := RootCmd.Execute(); err != nil {
		log.WithFields(log.Fields{
			"errmsg": err.Error(),
		}).Error("[rootcmd]")
		os.Exit(1)
	}
}
