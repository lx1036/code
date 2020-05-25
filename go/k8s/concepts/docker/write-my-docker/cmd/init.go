package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	InitCmd = &cobra.Command{
		Use:   "init",
		Short: "Init container process run user's process in container. Do not call it outside",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("init come on")
			_ = container.RunContainerInitProcess()
		},
	}
)

func init() {

}
